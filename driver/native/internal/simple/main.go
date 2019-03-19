package main

import (
	"context"

	"gopkg.in/bblfsh/sdk.v2/driver/native"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

type mockDriver struct{}

func (mockDriver) Start() error {
	return nil
}

func (mockDriver) Parse(ctx context.Context, src string) (nodes.Node, error) {
	if src == "die" {
		panic("died")
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
