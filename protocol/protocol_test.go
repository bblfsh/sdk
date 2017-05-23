package protocol_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/bblfsh/sdk/protocol"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type mockParser struct{}

func (p *mockParser) ParseUAST(req *protocol.ParseUASTRequest) *protocol.ParseUASTResponse {
	return &protocol.ParseUASTResponse{Status: protocol.Ok}
}

func (p *mockParser) Close() error {
	return nil
}

func TestInvalidParser(t *testing.T) {
	require := require.New(t)

	protocol.DefaultParser = nil
	lis, err := net.Listen("tcp", ":0")
	require.NoError(err)

	server := grpc.NewServer()
	protocol.RegisterProtocolServiceServer(
		server,
		protocol.NewProtocolServiceServer(),
	)

	go server.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTimeout(time.Second*2), grpc.WithInsecure())
	require.NoError(err)

	client := protocol.NewProtocolServiceClient(conn)

	ureq := &protocol.ParseUASTRequest{
		Content: "my source code",
	}
	uresp, err := client.ParseUAST(context.TODO(), ureq)
	require.NoError(err)
	require.Equal(protocol.Fatal, uresp.Status)

	server.GracefulStop()
}

func Example() {
	lis, err := net.Listen("tcp", ":0")
	checkError(err)

	// Use a mock parser on the server.
	protocol.DefaultParser = &mockParser{}

	server := grpc.NewServer()
	protocol.RegisterProtocolServiceServer(
		server,
		protocol.NewProtocolServiceServer(),
	)

	go server.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithTimeout(time.Second*2), grpc.WithInsecure())
	checkError(err)

	client := protocol.NewProtocolServiceClient(conn)

	req := &protocol.ParseUASTRequest{Content: "my source code"}
	fmt.Println("Sending ParseUAST for:", req.Content)
	resp, err := client.ParseUAST(context.TODO(), req)
	checkError(err)
	fmt.Println("Got response with status:", resp.Status.String())

	server.GracefulStop()

	//Output: Sending ParseUAST for: my source code
	// Got response with status: ok
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
