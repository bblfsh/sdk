package main

import (
	"context"
	"fmt"
	"os"

	"github.com/bblfsh/sdk/v3/driver/native"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

type mockDriver struct{}

func (mockDriver) Start() error {
	return nil
}

func (mockDriver) Parse(ctx context.Context, src string) (nodes.Node, error) {
	switch src {
	case "die":
		// just die, prints stack trace on stderr
		panic("died")
	case "print-and-die":
		// protocol runs on stdout, break it and then exit
		fmt.Println("crash command received")
		os.Exit(0)
	}
	return nodes.Object{
		"root": nodes.Object{
			"key": nodes.String(src),
		},
	}, nil
}

func (mockDriver) Close() error {
	return nil
}

func main() {
	native.Main(mockDriver{})
}
