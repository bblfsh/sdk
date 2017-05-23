package protocol

import (
	"golang.org/x/net/context"
)

type protocolServiceServer struct {
}

func NewProtocolServiceServer() *protocolServiceServer {
	return &protocolServiceServer{}
}
func (s *protocolServiceServer) ParseAST(ctx context.Context, in *ParseASTRequest) (result *ParseASTResponse, err error) {
	result = new(ParseASTResponse)
	result = ParseAST(in)
	return
}
func (s *protocolServiceServer) ParseUAST(ctx context.Context, in *ParseUASTRequest) (result *ParseUASTResponse, err error) {
	result = new(ParseUASTResponse)
	result = ParseUAST(in)
	return
}
