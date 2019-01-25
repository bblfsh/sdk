package main

import (
	"context"
	"time"

	"gopkg.in/bblfsh/sdk.v2/driver/native"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
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
