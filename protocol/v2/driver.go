package protocol

import (
	"bytes"
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/bblfsh/sdk.v2/driver"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes/nodesproto"
)

//go:generate protoc --proto_path=$GOPATH/src:. --gogo_out=plugins=grpc:. ./driver.proto

func RegisterDriver(srv *grpc.Server, d *driver.Driver) {
	RegisterDriverServer(srv, &driverServer{d: d})
}

var modeMap = map[Mode]driver.Mode{
	Mode_DefaultMode: driver.ModeDefault,
	Mode_Native:      driver.ModeNative,
	Mode_Annotated:   driver.ModeAnnotated,
	Mode_Semantic:    driver.ModeSemantic,
}

type driverServer struct {
	d *driver.Driver
}

func (s *driverServer) Parse(ctx context.Context, req *ParseRequest) (*ParseResponse, error) {
	mode, ok := modeMap[req.Mode]
	if !ok {
		return nil, fmt.Errorf("unsupported mode: %v", req.Mode)
	}
	var resp ParseResponse
	n, err := s.d.Parse(ctx, mode, req.Content)
	if e, ok := err.(*driver.ErrPartialParse); ok {
		n = e.AST
		for _, txt := range e.Errors {
			resp.Errors = append(resp.Errors, &ParseError{Text: txt})
		}
	} else if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	err = nodesproto.WriteTo(buf, n)
	if err != nil {
		return nil, err
	}
	resp.Uast = buf.Bytes()
	return &resp, nil
}
