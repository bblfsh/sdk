package driver

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"

	"github.com/bblfsh/sdk/v3/driver/manifest"
	"github.com/bblfsh/sdk/v3/uast/nodes"
)

// NewDriverFrom returns a new DriverModule instance based on
// the given pool of native drivers, language manifest and list of transformers.
func NewDriverFrom(ch chan Native, m *manifest.Manifest, t Transforms) (DriverModule, error) {
	if ch == nil || len(ch) == 0 {
		return nil, fmt.Errorf("no driver implementation")
	}

	if m == nil {
		return nil, fmt.Errorf("no manifest")
	}

	return &driverImpl{ch: ch, m: m, t: t, done: make(chan struct{})}, nil
}

// Driver implements a bblfsh driver, a driver is on charge of transforming a
// source code into an AST and a UAST. To transform the AST into a UAST, a
// `uast.ObjectToNode`` and a series of `tranformer.Transformer` are used.
//
// The `Parse` and `NativeParse` requests block the driver until the request is
// done. The communication with the native driver is based on a buffer channel of drivers,
// internally it allows number of concurrent requests equal to channel's size (by default - number of CPUs).
// Communication with single driver is synchronous and goes over stdin/stdout.
type driverImpl struct {
	ch chan Native

	m    *manifest.Manifest
	t    Transforms
	done chan struct{}
}

// Start gets each driver from the channel, starts the process
// and puts it back to the same buffer channel.
// If any of processes fail on start,
// the function tries to close all running drivers and return an error.
func (d *driverImpl) Start() error {
	for i := 0; i < len(d.ch); i++ {
		drv := <-d.ch
		if err := drv.Start(); err != nil {
			d.Close()
			return err
		}
		d.ch <- drv
	}
	return nil
}

// Close tries to close all idle drivers.
// If any of drivers fail on Close, the function returns the last received error.
func (d *driverImpl) Close() error {
	select {
	case <-d.done:
		return nil

	default:
		close(d.done)

		var last error
		for i := 0; i < len(d.ch); i++ {
			drv := <-d.ch
			if err := drv.Close(); err != nil {
				last = err
			}
		}
		return last
	}
}

// Parse process a protocol.ParseRequest, calling to the native driver. It a
// parser request is done to the internal native driver and the the returned
// native AST is transform to UAST.
func (d *driverImpl) Parse(rctx context.Context, src string, opts *ParseOptions) (nodes.Node, error) {
	sp, ctx := opentracing.StartSpanFromContext(rctx, "bblfsh.driver.Parse")
	defer sp.Finish()

	if opts == nil {
		opts = &ParseOptions{}
	}

	select {
	case <-rctx.Done():
		return nil, rctx.Err()

	case <-d.done:
		return nil, ErrDriverClosed.New()

	case drv := <-d.ch: // get a native driver or wait on available one.
		// put the driver back, when you're done.
		defer func() { d.ch <- drv }()

		ast, err := drv.Parse(ctx, src)
		if err != nil {
			if !ErrDriverFailure.Is(err) {
				// all other errors are considered syntax errors
				err = ErrSyntax.Wrap(err)
			} else {
				ast = nil
			}
			return ast, err
		}
		if opts.Language == "" {
			opts.Language = d.m.Language
		}

		ast, err = d.t.Do(ctx, opts.Mode, src, ast)
		if err != nil {
			err = ErrTransformFailure.Wrap(err)
		}
		return ast, err

	}
}

// Version returns driver version.
func (d *driverImpl) Version(ctx context.Context) (Version, error) {
	return Version{
		Version: d.m.Version,
		Build:   d.m.Build,
	}, nil
}

// Languages returns a single driver manifest for the language supported by the driver.
func (d *driverImpl) Languages(ctx context.Context) ([]manifest.Manifest, error) {
	return []manifest.Manifest{*d.m}, nil // TODO: clone
}
