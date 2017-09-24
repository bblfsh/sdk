package driver

import (
	"encoding/json"
	"time"

	"gopkg.in/bblfsh/sdk.v1/manifest"
	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/uast"
	"gopkg.in/bblfsh/sdk.v1/uast/transformer"
)

var (
	// ManifestLocation location of the manifest file.
	ManifestLocation = "/etc/" + manifest.Filename
)

// Driver implements a driver.
type Driver struct {
	NativeDriver
	m *manifest.Manifest
	o *uast.ObjectToNode
	t []transformer.Tranformer
}

// NewDriver returns a new Driver instance based on the given ObjectToNode and
// list of transformers.
func NewDriver(o *uast.ObjectToNode, t []transformer.Tranformer) (*Driver, error) {
	m, err := manifest.Load(ManifestLocation)
	if err != nil {
		return nil, err
	}

	return &Driver{m: m, o: o, t: t}, nil
}

// Parse process a protocol.ParseRequest, calling to the native driver. It calls
// to NativeDriver.NativeParse and transform the native AST to UAST using
// *uast.ObjectToNode and apply the transformers to  it.
func (d *Driver) Parse(req *protocol.ParseRequest) *protocol.ParseResponse {
	r := &protocol.ParseResponse{}

	start := time.Now()
	defer func() {
		r.Elapsed = time.Since(start)
	}()

	var ast interface{}
	r.Response, ast = d.doParse(req.Language, req.Content, req.Encoding)
	if r.Status == protocol.Fatal {
		return r
	}

	var err error
	r.UAST, err = d.convertToUAST(req.Content, req.Encoding, ast)
	if err != nil {
		r.Status = protocol.Fatal
		r.Errors = append(r.Errors, err.Error())
		return r
	}

	return r
}

func (d *Driver) convertToUAST(
	content string, encoding protocol.Encoding, ast interface{},
) (*uast.Node, error) {
	n, err := d.o.ToNode(ast)
	if err != nil {
		return nil, err
	}

	for _, t := range d.t {
		if err := t.Do(content, encoding, n); err != nil {
			return nil, err
		}
	}

	return n, nil
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
	if !d.isValidateLanguage(language, &r) {
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

func (d *Driver) isValidateLanguage(language string, r *protocol.Response) bool {
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
	r.Commit = d.m.Commit
	if d.m.Build != nil {
		r.Build = *d.m.Build
	}

	return r
}
