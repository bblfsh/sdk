package main

import (
	"context"
	"time"

	"github.com/bblfsh/sdk/v3/driver/native"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

type mockDriver struct{}

func (mockDriver) Start() error {
	time.Sleep(time.Second * 3)
	return nil
}

func (mockDriver) Parse(ctx context.Context, src string) (nodes.Node, error) {
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
