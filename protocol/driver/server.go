package driver

import (
	"io"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/protocol/jsonlines"

	"gopkg.in/src-d/go-errors.v0"
)

var (
	ErrDecodingRequest  = errors.NewKind("error decoding request")
	ErrEncodingResponse = errors.NewKind("error encoding response")

	noErrClose = errors.NewKind("clean close").New()
)

type Server struct {
	Parser

	In  io.Reader
	Out io.Writer

	err  error
	done chan struct{}
}

func (s *Server) Start() error {
	s.done = make(chan struct{})
	go func() {
		s.serve()
	}()

	return nil
}

func (s *Server) Wait() error {
	<-s.done
	return s.err
}

func (s *Server) serve() {
	enc := jsonlines.NewEncoder(s.Out)
	dec := jsonlines.NewDecoder(s.In)
	for {
		if err := s.process(enc, dec); err != nil {
			if err == noErrClose {
				break
			}

			s.err = err
			break
		}
	}

	close(s.done)
}

func (s *Server) process(enc jsonlines.Encoder, dec jsonlines.Decoder) error {
	req := &protocol.ParseRequest{}
	if err := dec.Decode(req); err != nil {
		if err == io.EOF {
			return noErrClose
		}

		return s.error(enc, ErrDecodingRequest.Wrap(err))
	}

	resp := s.Parser.Parse(req)
	if err := enc.Encode(resp); err != nil {
		return ErrEncodingResponse.Wrap(err)
	}

	return nil
}

func (s *Server) error(enc jsonlines.Encoder, err error) error {
	if err := enc.Encode(newFatalResponse(err)); err != nil {
		return ErrEncodingResponse.Wrap(err)
	}

	return nil
}

func newFatalResponse(err error) *protocol.ParseResponse {
	return &protocol.ParseResponse{
		Status: protocol.Fatal,
		Errors: []string{err.Error()},
	}
}
