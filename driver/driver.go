// Package driver contains all the logic to build a driver.
package driver

import (
	"bytes"
	"context"
	"fmt"

	"gopkg.in/bblfsh/sdk.v2/driver/manifest"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/transformer"
)

type Mode int

const (
	ModeNative = Mode(1 << iota)
	ModePreprocessed
	ModeAnnotated
	ModeSemantic
)

const ModeDefault = ModeSemantic

// Transforms describes a set of AST transformations this driver requires.
type Transforms struct {
	// Namespace for this language driver
	Namespace string
	// Preprocess transforms normalizes native AST.
	// It usually includes:
	//	* changing type key to uast.KeyType
	//	* changing token key to uast.KeyToken
	//	* restructure positional information
	Preprocess []transformer.Transformer
	// Normalize converts known AST structures to high-level AST representation (UAST).
	Normalize []transformer.Transformer
	// Annotations transforms annotates the tree with roles.
	Annotations []transformer.Transformer
	// Code transforms are applied directly after Native and provide a way
	// to extract more information from source files, fix positional info, etc.
	Code []transformer.CodeTransformer
}

// Do applies AST transformation pipeline for specified nodes.
func (t Transforms) Do(mode Mode, code string, nd nodes.Node) (nodes.Node, error) {
	if mode == 0 {
		mode = ModeDefault
	}
	if mode == ModeNative {
		return nd, nil
	}
	var err error

	runAll := func(list []transformer.Transformer) error {
		for _, t := range list {
			nd, err = t.Do(nd)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if err := runAll(t.Preprocess); err != nil {
		return nd, err
	}
	if mode >= ModeSemantic {
		if err := runAll(t.Normalize); err != nil {
			return nd, err
		}
	}
	if mode >= ModeAnnotated {
		if err := runAll(t.Annotations); err != nil {
			return nd, err
		}
	}

	for _, ct := range t.Code {
		t := ct.OnCode(code)
		nd, err = t.Do(nd)
		if err != nil {
			return nd, err
		}
	}
	if mode >= ModeSemantic && t.Namespace != "" {
		nd, err = transformer.DefaultNamespace(t.Namespace).Do(nd)
		if err != nil {
			return nd, err
		}
	}
	return nd, nil
}

// Module is an interface for a generic module instance.
type Module interface {
	Start() error
	Close() error
}

// Driver is an interface for a language driver that returns UAST.
type Driver interface {
	Parse(ctx context.Context, mode Mode, src string) (nodes.Node, error)
}

// DriverModule is an interface for a driver instance.
type DriverModule interface {
	Module
	Driver
	Manifest() (manifest.Manifest, error)
}

// Native is a base interface of a language driver that returns a native AST.
type Native interface {
	Module
	Parse(ctx context.Context, src string) (nodes.Node, error)
}

// ErrMulti joins multiple errors.
type ErrMulti struct {
	Errors []string
}

func (e ErrMulti) Error() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("partial parse:\n")
	for _, s := range e.Errors {
		buf.WriteString(s)
		buf.WriteString("\n")
	}
	return buf.String()
}

func MultiError(errs []string) error {
	return &ErrMulti{Errors: errs}
}

func PartialParse(ast nodes.Node, errs []string) error {
	return &ErrPartialParse{
		ErrMulti: ErrMulti{Errors: errs},
		AST:      ast,
	}
}

// ErrPartialParse is returned when driver was not able to parse the whole source file.
type ErrPartialParse struct {
	ErrMulti
	AST nodes.Node
}

// Driver implements a bblfsh driver, a driver is on charge of transforming a
// source code into an AST and a UAST. To transform the AST into a UAST, a
// `uast.ObjectToNode`` and a series of `tranformer.Transformer` are used.
//
// The `Parse` and `NativeParse` requests block the driver until the request is
// done, since the communication with the native driver is a single-channel
// synchronous communication over stdin/stdout.
type driverImpl struct {
	d Native

	m *manifest.Manifest
	t Transforms
}

// NewDriver returns a new Driver instance based on the given ObjectToNode and list of transformers.
func NewDriverFrom(d Native, m *manifest.Manifest, t Transforms) (DriverModule, error) {
	if d == nil {
		return nil, fmt.Errorf("no driver implementation")
	} else if m == nil {
		return nil, fmt.Errorf("no manifest")
	}
	return &driverImpl{d: d, m: m, t: t}, nil
}

func (d *driverImpl) Start() error {
	return d.d.Start()
}

func (d *driverImpl) Close() error {
	return d.d.Close()
}

// Parse process a protocol.ParseRequest, calling to the native driver. It a
// parser request is done to the internal native driver and the the returned
// native AST is transform to UAST.
func (d *driverImpl) Parse(ctx context.Context, mode Mode, src string) (nodes.Node, error) {
	ast, err := d.d.Parse(ctx, src)
	if err != nil {
		return nil, err
	}
	return d.t.Do(mode, src, ast)
}

// Manifest returns a driver manifest.
func (d *driverImpl) Manifest() (manifest.Manifest, error) {
	return *d.m, nil // TODO: clone
}
