package server

import (
	"net"

	"gopkg.in/bblfsh/sdk.v1/protocol"

	"github.com/Sirupsen/logrus"
	"google.golang.org/grpc"
	"gopkg.in/src-d/go-errors.v1"
)

var (
	ErrMissingParameter = errors.NewKind("missing configuration parameter: %s")
)

// Server is a common implementation of a gRPC server.
type Server struct {
	// Options list of grpc.ServerOption's.
	Options []grpc.ServerOption
	// Logger a logger to be used by the server.
	Logger *logrus.Logger

	*grpc.Server
}

// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each.
func (s *Server) Serve(listener net.Listener) error {
	if err := s.initialize(); err != nil {
		return err
	}

	defer func() {
		s.Logger.Info("grpc server ready")
	}()

	return s.Server.Serve(listener)
}

func (s *Server) initialize() error {
	if err := s.validate(); err != nil {
		return err
	}

	s.Server = grpc.NewServer(s.Options...)

	s.Logger.Debug("registering grpc service")
	protocol.RegisterProtocolServiceServer(
		s.Server,
		protocol.NewProtocolServiceServer(),
	)

	return nil
}

func (s *Server) validate() error {
	if s.Logger == nil {
		return ErrMissingParameter.New("logger")
	}

	return nil
}
