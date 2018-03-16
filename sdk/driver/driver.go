package driver

import (
	"encoding/json"
	"time"

	"gopkg.in/bblfsh/sdk.v1/manifest"
	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer"
)

type Mode int

const (
	ModeAST = Mode(iota)
	ModeRoles
)

// Transforms describes a set of AST transformations this driver requires.
type Transforms struct {
	// Native transforms normalizes native AST.
	// It usually includes:
	//	* changing type key to uast.KeyType
	//	* changing token key to uast.KeyToken
	//	* restructure positional information
	Native []transformer.Transformer
	// Code transforms are applied directly after Native and provide a way
	// to extract more information from source files, fix positional info, etc.
	Code []transformer.CodeTransformer
	// Roles transforms annotate the native AST tree with UAST roles.
	Roles []transformer.Transformer
}

// Do applies AST transformation pipeline for specified nodes.
func (t Transforms) Do(mode Mode, code string, nd uast.Node) (uast.Node, error) {
	var err error
	for _, t := range t.Native {
		nd, err = t.Do(nd)
		if err != nil {
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
	if mode >= ModeRoles {
		for _, t := range t.Roles {
			nd, err = t.Do(nd)
			if err != nil {
				return nd, err
			}
		}
	}
	return nd, nil
}

// Driver implements a bblfsh driver, a driver is on charge of transforming a
// source code into an AST and a UAST. To transform the AST into a UAST, a
// `uast.ObjectToNode`` and a series of `tranformer.Transformer` are used.
//
// The `Parse` and `NativeParse` requests block the driver until the request is
// done, since the communication with the native driver is a single-channel
// synchronous communication over stdin/stdout.
type Driver struct {
	NativeDriver

	m *manifest.Manifest
	t Transforms
}

// NewDriver returns a new Driver instance based on the given ObjectToNode and
// list of transformers.
func NewDriver(t Transforms) (*Driver, error) {
	m, err := manifest.Load(ManifestLocation)
	if err != nil {
		return nil, err
	}

	return &Driver{m: m, t: t}, nil
}

// Parse process a protocol.ParseRequest, calling to the native driver. It a
// parser request is done to the internal native driver and the the returned
// native AST is transform to UAST.
func (d *Driver) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	r := &protocol.ParseResponse{}

	start := time.Now()
	defer func() {
		r.Elapsed = time.Since(start)
	}()

	var ast interface{}
	r.Response, ast = d.doParse(req.Language, req.Content, req.Encoding)

	if r.Language == "" {
		r.Language = d.m.Language
	}

	if r.Filename == "" {
		r.Filename = req.Filename
	}

	if r.Status == protocol.Fatal {
		return r
	}

	addErr := func(err error) {
		r.Status = protocol.Fatal
		r.Errors = append(r.Errors, err.Error())
	}

	nd, err := uast.ToNode(ast)
	if err != nil {
		addErr(err)
		return r
	}

	code := req.Content
	code, err = Encoding(req.Encoding).Decode(code)
	if err != nil {
		addErr(err)
		return r
	}
	nd, err = d.t.Do(ModeRoles, code, nd)
	if err != nil {
		addErr(err)
		return r
	}
	r.UAST, err = protocol.ToNode(nd)
	if err != nil {
		addErr(err)
		return r
	}
	return r
}

// NativeParse sends a request to the native driver and returns its response.
func (d *Driver) NativeParse(req *protocol.NativeParseRequest) *protocol.NativeParseResponse {
	r := &protocol.NativeParseResponse{}

	start := time.Now()
	defer func() {
		r.Elapsed = time.Since(start)
	}()

	var ast interface{}
	r.Response, ast = d.doParse(req.Language, req.Content, req.Encoding)

	if r.Language == "" {
		r.Language = d.m.Language
	}

	if r.Status == protocol.Fatal {
		return r
	}

	js, err := json.Marshal(&ast)
	if err != nil {
		r.Errors = append(r.Errors, err.Error())
	}

	r.AST = string(js)
	return r
}

func (d *Driver) doParse(language, content string, encoding protocol.Encoding) (
	r protocol.Response, ast interface{},
) {
	if !d.isValidLanguage(language, &r) {
		return r, nil
	}

	nr := d.NativeDriver.Parse(&InternalParseRequest{
		Content:  content,
		Encoding: Encoding(encoding),
	})

	r.Status = protocol.Status(nr.Status)
	r.Errors = nr.Errors

	ast = nr.AST
	return
}

func (d *Driver) isValidLanguage(language string, r *protocol.Response) bool {
	if language == d.m.Language {
		return true
	}

	r.Status = protocol.Fatal
	r.Errors = append(r.Errors,
		ErrUnsupportedLanguage.New(language, d.m.Language).Error(),
	)

	return false
}

// Version handles a VersionRequest including information from the manifest.
func (d *Driver) Version(req *protocol.VersionRequest) *protocol.VersionResponse {
	r := &protocol.VersionResponse{}

	r.Version = d.m.Version
	if d.m.Build != nil {
		r.Build = *d.m.Build
	}

	return r
}
