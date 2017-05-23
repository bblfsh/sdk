package protocol

import (
	"golang.org/x/net/context"
)

type protocolServiceServer struct {
}

func NewProtocolServiceServer() *protocolServiceServer {
	return &protocolServiceServer{}
}
func (s *protocolServiceServer) ParseUAST(ctx context.Context, in *ParseUASTRequest) (result *ParseUASTResponse, err error) {
	result = new(ParseUASTResponse)
	result = ParseUAST(in)
	return
}
