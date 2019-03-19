package main

import (
	"context"

	"gopkg.in/bblfsh/sdk.v2/driver/native"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
)

type mockDriver struct{}

func (mockDriver) Start() error {
	panic("died")
	return nil
}

func (mockDriver) Parse(ctx context.Context, src string) (nodes.Node, error) {
	panic("unreachable")
}

func (mockDriver) Close() error {
	return nil
}

func main() {
	native.Main(mockDriver{})
}
