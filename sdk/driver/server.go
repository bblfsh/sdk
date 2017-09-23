package driver

import (
	"flag"
	"net"
	"os"

	"github.com/Sirupsen/logrus"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/sdk/server"
)

var (
	network *string
	address *string
	verbose *string
)

// Server is a grpc server for the communication with the driver.
type Server struct {
	server.Server
	Driver
}

// Start executes the binary driver and start to listen in the network and
// address defined by the args.
func (s *Server) Start() error {
	if err := s.initialize(); err != nil {
		return err
	}

	s.Logger.Debugf("executing native binary %q ...", s.Binary)
	if err := s.Driver.Start(); err != nil {
		return err
	}

	l, err := net.Listen(*network, *address)
	if err != nil {
		return err
	}

	s.Logger.Infof("server listening in %s (%s)", *address, *network)

	protocol.DefaultParser = s
	protocol.DefaultVersioner = s

	return s.Serve(l)
}

func (s *Server) initialize() error {
	s.initializeFlags()
	if err := s.initializeLogger(); err != nil {
		return err
	}

	s.Logger.Infof("version: %q, commit: %q, build: %q",
		s.Driver.Version,
		s.Driver.Commit,
		s.Driver.Build,
	)

	return nil
}

func (s *Server) initializeFlags() {
	const (
		defaultNetwork = "tcp"
		defaultAddress = "localhost:9432"
		defaultVerbose = "info"
	)

	cmd := flag.NewFlagSet("server", flag.ExitOnError)
	network = cmd.String("network", defaultNetwork, "network type: tcp, tcp4, tcp6, unix or unixpacket.")
	address = cmd.String("address", defaultAddress, "address to listen")
	verbose = cmd.String("verbose", defaultVerbose, "verbose level: panic, fatal, error, warning, info, debug.")

	cmd.Parse(os.Args[1:])
}

func (s *Server) initializeLogger() error {
	s.Logger = logrus.New()
	if *verbose != "" {
		level, err := logrus.ParseLevel(*verbose)
		if err != nil {
			return err
		}

		s.Logger.Level = level
	}

	return nil
}

// Parse handles a ParserRequest.
func (s *Server) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	resp := s.Driver.Parse(req)
	s.Logger.Infof("request processed content %d bytes, encoding %s, status %s in %s",
		len(req.Content), req.Encoding, resp.Status, resp.Elapsed)

	return resp
}

// Version handles a VersionRequest.
func (s *Server) Version(req *protocol.VersionRequest) *protocol.VersionResponse {
	return &protocol.VersionResponse{
		Version: s.Driver.Version,
		Build:   s.Driver.Build,
		Commit:  s.Driver.Commit,
	}
}
