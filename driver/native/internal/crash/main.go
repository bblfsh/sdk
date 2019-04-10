package main

import (
	"context"

	"github.com/bblfsh/sdk/v3/driver/native"
	"github.com/bblfsh/sdk/v3/uast/nodes"
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
