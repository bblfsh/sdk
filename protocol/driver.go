package protocol

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	serrors "gopkg.in/src-d/go-errors.v1"

	"github.com/bblfsh/sdk/v3/driver"
	"github.com/bblfsh/sdk/v3/driver/manifest"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"github.com/bblfsh/sdk/v3/uast/nodes/nodesproto"
)

//go:generate protoc --proto_path=$GOPATH/src:. --gogo_out=plugins=grpc:. ./driver.proto
// TODO(dennwc): bug in gogo protobuf generator; fix upstream
//go:generate sed -i "s/dAtA\\[\\:m\\.Size/dAtA\\[\\:m\\.ProtoSize/g" driver.pb.go

const (
	mb = 1 << 20

	// DefaultGRPCMaxMessageBytes is maximum msg size for gRPC.
	DefaultGRPCMaxMessageBytes = 100 * mb
)

// ServerOptions returns a set of common options that should be used in bblfsh server.
//
// It automatically enables OpenTrace if a global tracer is set.
func ServerOptions() []grpc.ServerOption {
	opts := []grpc.ServerOption{
		grpc.MaxSendMsgSize(DefaultGRPCMaxMessageBytes),
		grpc.MaxRecvMsgSize(DefaultGRPCMaxMessageBytes),
	}
	tracer := opentracing.GlobalTracer()
	if _, ok := tracer.(opentracing.NoopTracer); ok {
		return opts
	}
	opts = append(opts,
		grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(tracer)),
		grpc.StreamInterceptor(otgrpc.OpenTracingStreamServerInterceptor(tracer)),
	)
	return opts
}

// DialOptions returns a set of common options that should be used when dialing bblfsh server.
//
// It automatically enables OpenTrace if a global tracer is set.
func DialOptions() []grpc.DialOption {
	opts := []grpc.DialOption{grpc.WithDefaultCallOptions(
		grpc.MaxCallSendMsgSize(DefaultGRPCMaxMessageBytes),
		grpc.MaxCallRecvMsgSize(DefaultGRPCMaxMessageBytes),
	)}
	tracer := opentracing.GlobalTracer()
	if _, ok := tracer.(opentracing.NoopTracer); ok {
		return opts
	}
	opts = append(opts,
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(tracer)),
		grpc.WithStreamInterceptor(otgrpc.OpenTracingStreamClientInterceptor(tracer)),
	)
	return opts
}

// RegisterDriver registers a v2 driver server on a given gRPC server.
func RegisterDriver(srv *grpc.Server, d driver.Driver) {
	s := &driverServer{d: d}
	RegisterDriverServer(srv, s)
	RegisterDriverHostServer(srv, s)
}

// AsDriver creates a v2 driver client for a given gRPC client.
func AsDriver(cc *grpc.ClientConn) driver.Driver {
	return &client{
		c: NewDriverClient(cc),
		h: NewDriverHostClient(cc),
	}
}

func toParseErrors(err error) []*ParseError {
	if e, ok := err.(*driver.ErrMulti); ok {
		errs := make([]*ParseError, 0, len(e.Errors))
		for _, e := range e.Errors {
			errs = append(errs, &ParseError{Text: e.Error()})
		}
		return errs
	}
	return []*ParseError{
		{Text: err.Error()},
	}
}

// newGRPCError creates a new gRPC error with a specified code, message and optional details.
// The function will panic if any error details fail to encode.
func newGRPCError(code codes.Code, cause error, details ...proto.Message) error {
	st, err := status.New(code, cause.Error()).WithDetails(details...)
	if err != nil {
		panic(err)
	}
	return st.Err()
}

type driverServer struct {
	d driver.Driver
}

// toGRPCError converts an error to gRPC equivalent.
// Some errors may be silenced and added to resp instead (e.g. syntax errors).
func toGRPCError(resp *ParseResponse, err error) error {
	e, ok := err.(*serrors.Error)
	if !ok {
		return err
	}
	cause := e.Cause()
	if cause == nil {
		cause = e // if no cause is available, use error's message instead
	}
	switch {
	case driver.ErrDriverFailure.Is(err):
		return newGRPCError(codes.Internal, cause, &ErrorDetails{
			Reason: &ErrorDetails_DriverFailure{DriverFailure: true},
		})
	case driver.ErrTransformFailure.Is(err):
		return newGRPCError(codes.FailedPrecondition, cause, &ErrorDetails{
			Reason: &ErrorDetails_TransformFailure{TransformFailure: true},
		})
	case driver.ErrModeNotSupported.Is(err):
		return newGRPCError(codes.InvalidArgument, cause, &ErrorDetails{
			Reason: &ErrorDetails_UnsupportedTransformMode{UnsupportedTransformMode: true},
		})
	case driver.ErrLanguageDetection.Is(err):
		return newGRPCError(codes.InvalidArgument, cause, &ErrorDetails{
			Reason: &ErrorDetails_CannotDetectLanguage{CannotDetectLanguage: true},
		})
	case driver.ErrUnknownEncoding.Is(err):
		return newGRPCError(codes.InvalidArgument, cause, &ErrorDetails{
			Reason: &ErrorDetails_InvalidFileEncoding{InvalidFileEncoding: true},
		})
	case driver.ErrSyntax.Is(err):
		// partial parse or syntax error; we will send an OK status code, but will fill Errors field
		resp.Errors = toParseErrors(cause)
		return nil
	}
	if e, ok := err.(*driver.ErrMissingDriver); ok {
		return newGRPCError(codes.InvalidArgument, cause, &ErrorDetails{
			Reason: &ErrorDetails_UnsupportedLanguage{UnsupportedLanguage: e.Language},
		})
	}
	return err // unknown error
}

// Parse implements DriverServer.
func (s *driverServer) Parse(rctx context.Context, req *ParseRequest) (*ParseResponse, error) {
	sp, ctx := opentracing.StartSpanFromContext(rctx, "bblfsh.server.Parse")
	defer sp.Finish()

	opts := &driver.ParseOptions{
		Mode:     driver.Mode(req.Mode),
		Language: req.Language,
		Filename: req.Filename,
	}
	var resp ParseResponse
	n, err := s.d.Parse(ctx, req.Content, opts)
	resp.Language = opts.Language // can be set during the call
	err = toGRPCError(&resp, err)
	if err != nil {
		return nil, err
	}

	dsp, _ := opentracing.StartSpanFromContext(ctx, "uast.Encode")
	defer dsp.Finish()

	buf := bytes.NewBuffer(nil)
	err = nodesproto.WriteTo(buf, n)
	if err != nil {
		return nil, err // unknown error = server failure
	}
	resp.Uast = buf.Bytes()
	return &resp, nil
}

func (s *driverServer) ServerVersion(rctx context.Context, _ *VersionRequest) (*VersionResponse, error) {
	sp, ctx := opentracing.StartSpanFromContext(rctx, "bblfsh.server.Parse")
	defer sp.Finish()

	resp, err := s.d.Version(ctx)
	if err != nil {
		return nil, err
	}
	vers := Version{Version: resp.Version, Build: resp.Build}
	return &VersionResponse{Version: &vers}, nil
}

func (s *driverServer) SupportedLanguages(rctx context.Context, _ *SupportedLanguagesRequest) (*SupportedLanguagesResponse, error) {
	sp, ctx := opentracing.StartSpanFromContext(rctx, "bblfsh.server.Parse")
	defer sp.Finish()

	resp, err := s.d.Languages(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*Manifest, 0, len(resp))
	for _, m := range resp {
		out = append(out, NewManifest(&m))
	}
	return &SupportedLanguagesResponse{Languages: out}, nil
}

type client struct {
	c DriverClient
	h DriverHostClient
}

// fromGRPCError extract error details from gRPC error codes and error details and converts it to a native bblfsh error.
func fromGRPCError(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return err
	}
	// prefer detailed errors
	for _, d := range s.Details() {
		d, ok := d.(*ErrorDetails)
		if !ok {
			continue
		}
		switch r := d.Reason.(type) {
		case *ErrorDetails_UnsupportedLanguage:
			// special error type - return directly
			return &driver.ErrMissingDriver{Language: r.UnsupportedLanguage}
		case *ErrorDetails_InvalidFileEncoding:
			if r.InvalidFileEncoding {
				return driver.ErrUnknownEncoding.New()
			}
		case *ErrorDetails_CannotDetectLanguage:
			if r.CannotDetectLanguage {
				return driver.ErrLanguageDetection.New()
			}
		case *ErrorDetails_UnsupportedTransformMode:
			if r.UnsupportedTransformMode {
				return driver.ErrModeNotSupported.New()
			}
		case *ErrorDetails_TransformFailure:
			if r.TransformFailure {
				return driver.ErrTransformFailure.Wrap(errors.New(s.Message()))
			}
		case *ErrorDetails_DriverFailure:
			if r.DriverFailure {
				return driver.ErrDriverFailure.Wrap(errors.New(s.Message()))
			}
		}
	}
	// fallback to detection based on the status code
	var kind *serrors.Kind
	switch s.Code() {
	case codes.Internal:
		kind = driver.ErrDriverFailure
	case codes.FailedPrecondition:
		kind = driver.ErrTransformFailure
	case codes.InvalidArgument:
		kind = driver.ErrModeNotSupported
	}
	if kind != nil {
		return kind.Wrap(errors.New(s.Message()))
	}
	return err
}

// Parse implements DriverClient.
func (c *client) Parse(rctx context.Context, src string, opts *driver.ParseOptions) (nodes.Node, error) {
	sp, ctx := opentracing.StartSpanFromContext(rctx, "bblfsh.client.Parse")
	defer sp.Finish()

	req := &ParseRequest{Content: src}
	if opts != nil {
		req.Mode = Mode(opts.Mode)
		req.Language = opts.Language
		req.Filename = opts.Filename
	}
	resp, err := c.c.Parse(ctx, req)
	err = fromGRPCError(err)
	if err != nil {
		return nil, err // server or network error
	}
	if opts != nil && opts.Language == "" {
		opts.Language = resp.Language
	}

	dsp, _ := opentracing.StartSpanFromContext(ctx, "uast.Decode")
	defer dsp.Finish()

	// it may be still a parsing error
	return resp.Nodes()
}

func (m *ParseResponse) Nodes() (nodes.Node, error) {
	ast, err := nodesproto.ReadTree(bytes.NewReader(m.Uast))
	if err != nil {
		return nil, err
	}
	if len(m.Errors) != 0 {
		var errs []error
		for _, e := range m.Errors {
			errs = append(errs, errors.New(e.Text))
		}
		// syntax error or partial parse - return both UAST and an error
		err = driver.ErrSyntax.Wrap(driver.JoinErrors(errs))
	}
	return ast, err
}

// Version implements DriverHostClient.
func (c *client) Version(rctx context.Context) (driver.Version, error) {
	sp, ctx := opentracing.StartSpanFromContext(rctx, "bblfsh.client.Version")
	defer sp.Finish()

	resp, err := c.h.ServerVersion(ctx, &VersionRequest{})
	if err != nil {
		return driver.Version{}, err
	}
	return driver.Version{Version: resp.Version.Version, Build: resp.Version.Build}, nil
}

// Languages implements DriverHostClient.
func (c *client) Languages(rctx context.Context) ([]manifest.Manifest, error) {
	sp, ctx := opentracing.StartSpanFromContext(rctx, "bblfsh.client.Languages")
	defer sp.Finish()

	resp, err := c.h.SupportedLanguages(ctx, &SupportedLanguagesRequest{})
	if err != nil {
		return nil, err
	}
	out := make([]manifest.Manifest, len(resp.Languages))
	for i, m := range resp.Languages {
		m.toNative(&out[i])
	}
	return out, nil
}

// NewManifest converts driver manifest to the corresponding protocol message.
func NewManifest(m *manifest.Manifest) *Manifest {
	dm := &Manifest{
		Name:     m.Name,
		Language: m.Language,
		Aliases:  m.Aliases,
		Features: make([]string, 0, len(m.Features)),
	}
	if m.Version != "" || !m.Build.IsZero() {
		dm.Version = &Version{
			Version: m.Version,
			Build:   m.Build,
		}
	}
	switch m.Status {
	case manifest.Inactive:
		dm.Status = DevelopmentStatus_Inactive
	case manifest.Planning:
		dm.Status = DevelopmentStatus_Planning
	case manifest.PreAlpha:
		dm.Status = DevelopmentStatus_PreAlpha
	case manifest.Alpha:
		dm.Status = DevelopmentStatus_Alpha
	case manifest.Beta:
		dm.Status = DevelopmentStatus_Beta
	case manifest.Stable:
		dm.Status = DevelopmentStatus_Stable
	case manifest.Mature:
		dm.Status = DevelopmentStatus_Mature
	default:
		st, _ := strconv.Atoi(string(m.Status))
		dm.Status = DevelopmentStatus(st)
	}
	for _, f := range m.Features {
		dm.Features = append(dm.Features, strings.ToLower(string(f)))
	}
	return dm
}

// ToNative converts the manifest message to the driver manifest used by the SDK.
func (m *Manifest) toNative(dm *manifest.Manifest) {
	dm.Name = m.Name
	dm.Language = m.Language
	dm.Aliases = m.Aliases
	dm.Features = make([]manifest.Feature, 0, len(m.Features))
	if m.Version != nil {
		dm.Version = m.Version.Version
		dm.Build = m.Version.Build
	}
	switch m.Status {
	case DevelopmentStatus_Inactive:
		dm.Status = manifest.Inactive
	case DevelopmentStatus_Planning:
		dm.Status = manifest.Planning
	case DevelopmentStatus_PreAlpha:
		dm.Status = manifest.PreAlpha
	case DevelopmentStatus_Alpha:
		dm.Status = manifest.Alpha
	case DevelopmentStatus_Beta:
		dm.Status = manifest.Beta
	case DevelopmentStatus_Stable:
		dm.Status = manifest.Stable
	case DevelopmentStatus_Mature:
		dm.Status = manifest.Mature
	default:
		dm.Status = manifest.DevelopmentStatus(strconv.Itoa(int(m.Status)))
	}
	for _, f := range m.Features {
		dm.Features = append(dm.Features, manifest.Feature(strings.ToLower(f)))
	}
}

// ToNative converts the manifest message to the driver manifest used by the SDK.
func (m *Manifest) ToNative() *manifest.Manifest {
	var dm manifest.Manifest
	m.toNative(&dm)
	return &dm
}
