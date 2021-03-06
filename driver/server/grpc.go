package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bblfsh/sdk/v3/driver"
	"github.com/bblfsh/sdk/v3/driver/manifest"
	protocol2 "github.com/bblfsh/sdk/v3/protocol"
	uast1 "github.com/bblfsh/sdk/v3/protocol/v1"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	"google.golang.org/grpc"
	protocol1 "gopkg.in/bblfsh/sdk.v1/protocol"
)

// NewGRPCServer creates a gRPC server instance that dispatches requests to a provided driver.
//
// It will automatically include default server options for bblfsh protocol.
func NewGRPCServer(drv driver.DriverModule, opts ...grpc.ServerOption) *grpc.Server {
	opts = append(opts, protocol2.ServerOptions()...)
	return NewGRPCServerCustom(drv, opts...)
}

// NewGRPCServerCustom is the same as NewGRPCServer, but it won't include any options except the ones that were passed.
func NewGRPCServerCustom(drv driver.DriverModule, opts ...grpc.ServerOption) *grpc.Server {
	srv := grpc.NewServer(opts...)

	protocol1.DefaultService = service{drv}
	protocol1.RegisterProtocolServiceServer(
		srv,
		protocol1.NewProtocolServiceServer(),
	)
	protocol2.RegisterDriver(srv, drv)

	return srv
}

type service struct {
	d driver.DriverModule
}

func errResp(err error) protocol1.Response {
	return protocol1.Response{Status: protocol1.Fatal, Errors: []string{err.Error()}}
}

func newDriverManifest(manifest *manifest.Manifest) protocol1.DriverManifest {
	features := make([]string, len(manifest.Features))
	for i, feature := range manifest.Features {
		features[i] = string(feature)
	}
	return protocol1.DriverManifest{
		Name:     manifest.Name,
		Language: manifest.Language,
		Version:  manifest.Version,
		Status:   string(manifest.Status),
		Features: features,
	}
}

func containsLang(lang string, list []manifest.Manifest) bool {
	for _, m := range list {
		if m.Language == lang {
			return true
		}
		for _, l := range m.Aliases {
			if l == lang {
				return true
			}
		}
	}
	return false
}

// SupportedLanguages implements protocol1.Service.
func (s service) SupportedLanguages(_ *protocol1.SupportedLanguagesRequest) *protocol1.SupportedLanguagesResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	list, err := s.d.Languages(ctx)
	if err != nil {
		return &protocol1.SupportedLanguagesResponse{Response: errResp(err)}
	}
	resp := &protocol1.SupportedLanguagesResponse{
		Response: protocol1.Response{Status: protocol1.Ok},
	}
	for _, m := range list {
		resp.Languages = append(resp.Languages, newDriverManifest(&m))
	}
	return resp
}

func (s service) parse(mode driver.Mode, req *protocol1.ParseRequest) (nodes.Node, protocol1.Response) {
	start := time.Now()
	ctx := context.Background()
	if req.Timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}
	list, err := s.d.Languages(ctx)
	if err != nil {
		r := errResp(err)
		r.Elapsed = time.Since(start)
		return nil, r
	}
	if !containsLang(req.Language, list) {
		r := errResp(ErrUnsupportedLanguage.New(req.Language))
		r.Elapsed = time.Since(start)
		return nil, r
	}
	ast, err := s.d.Parse(ctx, req.Content, &driver.ParseOptions{
		Mode:     mode,
		Language: req.Language,
		Filename: req.Filename,
	})
	dt := time.Since(start)
	var r protocol1.Response
	if err != nil {
		r = errResp(err)
	} else {
		r = protocol1.Response{Status: protocol1.Ok}
	}
	r.Elapsed = dt
	return ast, r
}

// Parse implements protocol1.Service.
func (s service) Parse(req *protocol1.ParseRequest) *protocol1.ParseResponse {
	ast, resp := s.parse(driver.ModeAnnotated, req)
	if resp.Status != protocol1.Ok {
		return &protocol1.ParseResponse{Response: resp}
	}
	nd, err := uast1.ToNode(ast)
	if err != nil {
		r := errResp(err)
		r.Elapsed = resp.Elapsed
		return &protocol1.ParseResponse{Response: r}
	}
	return &protocol1.ParseResponse{
		Response: resp,
		Language: req.Language,
		Filename: req.Filename,
		UAST:     nd,
	}
}

// NativeParse implements protocol1.Service.
func (s service) NativeParse(req *protocol1.NativeParseRequest) *protocol1.NativeParseResponse {
	ast, resp := s.parse(driver.ModeNative, (*protocol1.ParseRequest)(req))
	if resp.Status != protocol1.Ok {
		return &protocol1.NativeParseResponse{Response: resp}
	}
	data, err := json.Marshal(ast)
	if err != nil {
		r := errResp(err)
		r.Elapsed = resp.Elapsed
		return &protocol1.NativeParseResponse{Response: r}
	}
	return &protocol1.NativeParseResponse{
		Response: resp,
		Language: req.Language,
		AST:      string(data),
	}
}

// Version implements protocol1.Service.
func (s service) Version(req *protocol1.VersionRequest) *protocol1.VersionResponse {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	m, err := s.d.Version(ctx)
	if err != nil {
		return &protocol1.VersionResponse{Response: errResp(err)}
	}
	return &protocol1.VersionResponse{
		Version: m.Version,
		Build:   m.Build,
	}
}
