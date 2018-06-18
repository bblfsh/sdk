package server

import (
	"context"
	"encoding/json"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/bblfsh/sdk.v2/driver"
	"gopkg.in/bblfsh/sdk.v2/protocol"
	protocol2 "gopkg.in/bblfsh/sdk.v2/protocol/v2"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

// NewGRPCServer creates a gRPC server.
func NewGRPCServer(drv *driver.Driver, opts ...grpc.ServerOption) *GRPCServer {
	return &GRPCServer{drv: drv, Options: opts}
}

// GRPCServer is a common implementation of a gRPC server.
type GRPCServer struct {
	// Options list of grpc.ServerOption's.
	Options []grpc.ServerOption

	drv *driver.Driver
	*grpc.Server
}

// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each.
func (s *GRPCServer) Serve(listener net.Listener) error {
	if err := s.initialize(); err != nil {
		return err
	}

	defer func() {
		logrus.Infof("grpc server ready")
	}()

	return s.Server.Serve(listener)
}

func (s *GRPCServer) initialize() error {
	s.Server = grpc.NewServer(s.Options...)

	logrus.Debugf("registering grpc service")

	protocol.DefaultService = service{s.drv}
	protocol.RegisterProtocolServiceServer(
		s.Server,
		protocol.NewProtocolServiceServer(),
	)
	protocol2.RegisterDriver(s.Server, s.drv)

	return nil
}

type service struct {
	d *driver.Driver
}

func (s service) Start() error {
	return s.d.Start()
}

func (s service) Stop() error {
	return s.d.Stop()
}

func errResp(err error) protocol.Response {
	return protocol.Response{Status: protocol.Fatal, Errors: []string{err.Error()}}
}

func (s service) parse(mode driver.Mode, req *protocol.ParseRequest) (nodes.Node, protocol.Response) {
	start := time.Now()
	m := s.d.Manifest()
	if req.Language != m.Language {
		r := errResp(ErrUnsupportedLanguage.New(req.Language))
		r.Elapsed = time.Since(start)
		return nil, r
	}
	ctx := context.Background()
	if req.Timeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}
	ast, err := s.d.Parse(ctx, mode, req.Content)
	dt := time.Since(start)
	var r protocol.Response
	if err != nil {
		r = errResp(err)
	} else {
		r = protocol.Response{Status: protocol.Ok}
	}
	r.Elapsed = dt
	return ast, r
}

func (s service) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	ast, resp := s.parse(driver.ModeSemantic, req)
	if resp.Status != protocol.Ok {
		return &protocol.ParseResponse{Response: resp}
	}
	nd, err := protocol.ToNode(ast)
	if err != nil {
		r := errResp(err)
		r.Elapsed = resp.Elapsed
		return &protocol.ParseResponse{Response: r}
	}
	return &protocol.ParseResponse{
		Response: resp,
		Language: req.Language,
		Filename: req.Filename,
		UAST:     nd,
	}
}

func (s service) NativeParse(req *protocol.NativeParseRequest) *protocol.NativeParseResponse {
	ast, resp := s.parse(driver.ModeNative, (*protocol.ParseRequest)(req))
	if resp.Status != protocol.Ok {
		return &protocol.NativeParseResponse{Response: resp}
	}
	data, err := json.Marshal(ast)
	if err != nil {
		r := errResp(err)
		r.Elapsed = resp.Elapsed
		return &protocol.NativeParseResponse{Response: r}
	}
	return &protocol.NativeParseResponse{
		Response: resp,
		Language: req.Language,
		AST:      string(data),
	}
}

func (s service) Version(req *protocol.VersionRequest) *protocol.VersionResponse {
	m := s.d.Manifest()

	r := &protocol.VersionResponse{
		Version: m.Version,
	}
	if m.Build != nil {
		r.Build = *m.Build
	}
	return r
}
