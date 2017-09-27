package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"google.golang.org/grpc"

	"gopkg.in/bblfsh/client-go.v1"
	"gopkg.in/bblfsh/sdk.v1/protocol"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/require"
)

const (
	DefaultFixtureLocation = "fixtures"
)

var Suite *suite

func init() {
	Suite = &suite{
		Endpoint: os.Getenv(Endpoint),
		Language: os.Getenv(Language),
		Fixtures: filepath.Join(os.Getenv(DriverPath), DefaultFixtureLocation),
	}
}

type suite struct {
	// Language of the driver being test.
	Language string
	// Endpoint of the grpc server to test.
	Endpoint string
	// Fixture to use against the driver
	Fixtures string

	c *bblfsh.BblfshClient
	g protocol.ProtocolServiceClient
}

func (s *suite) SetUpTest(t *testing.T) {
	if s.Endpoint == "" || s.Language == "" {
		t.Skip("please use `bblfsh-sdk-tools test`")
	}

	r := require.New(t)
	client, err := bblfsh.NewBblfshClient(s.Endpoint)
	r.Nil(err)

	s.c = client

	// TODO: use client-go as soon NativeParse request is availabe on it.
	conn, err := grpc.Dial(s.Endpoint, grpc.WithTimeout(time.Second*2), grpc.WithInsecure())
	r.Nil(err)

	s.g = protocol.NewProtocolServiceClient(conn)
}

func (s *suite) TestParse(t *testing.T) {
	files, err := filepath.Glob(fmt.Sprintf("%s/*.source", s.Fixtures))
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		filename := removeExtension(f)
		t.Run(filepath.Base(filename), func(t *testing.T) {
			s.doTestParse(t, filename)
		})
	}
}

func (s *suite) TestNativeParse(t *testing.T) {
	files, err := filepath.Glob(fmt.Sprintf("%s/*.source", s.Fixtures))
	if err != nil {
		panic(err)
	}

	for _, f := range files {
		filename := removeExtension(f)
		t.Run(filepath.Base(filename), func(t *testing.T) {
			s.doTestNativeParse(t, filename)
		})
	}
}

func (s *suite) doTestParse(t *testing.T, filename string) {
	r := require.New(t)

	source := getSourceCode(r, filename)
	req := s.c.NewParseRequest().Language(s.Language).Content(source)

	res, err := req.Do()
	r.Nil(err)

	expected := getUAST(r, filename)
	EqualText(r, expected, res.String())
}

func (s *suite) doTestNativeParse(t *testing.T, filename string) {
	r := require.New(t)

	source := getSourceCode(r, filename)
	req := &protocol.NativeParseRequest{
		Language: s.Language,
		Content:  source,
	}

	res, err := s.g.NativeParse(context.Background(), req)
	r.Nil(err)

	expected := getAST(r, filename)

	EqualText(r, expected, NativeParseResponseToString(r, res))
}

func NativeParseResponseToString(r *require.Assertions, res *protocol.NativeParseResponse) string {
	var s struct {
		Status string      `json:"status"`
		Errors []string    `json:"errors"`
		AST    interface{} `json:"ast"`
	}

	s.Status = strings.ToLower(res.Status.String())
	s.Errors = res.Errors
	if len(s.Errors) == 0 {
		s.Errors = make([]string, 0)
	}

	err := json.Unmarshal([]byte(res.AST), &s.AST)
	r.Nil(err)

	buf := bytes.NewBuffer(nil)
	e := json.NewEncoder(buf)
	e.SetIndent("", "    ")
	e.SetEscapeHTML(false)

	err = e.Encode(s)
	r.Nil(err)

	return buf.String()
}

func EqualText(r *require.Assertions, expected, actual string) {
	diff := difflib.ContextDiff{
		A:        difflib.SplitLines(expected),
		B:        difflib.SplitLines(actual),
		FromFile: "expected",
		ToFile:   "actual",
		Context:  3,
		Eol:      "\n",
	}

	patch, err := difflib.GetContextDiffString(diff)
	r.Nil(err)

	if patch != "" {
		r.Fail("response doesn't match", patch)
	}
}

func getSourceCode(r *require.Assertions, filename string) string {
	return getFileContent(r, filename, "source")
}

func getUAST(r *require.Assertions, filename string) string {
	return getFileContent(r, filename, "uast")
}

func getAST(r *require.Assertions, filename string) string {
	return getFileContent(r, filename, "native")
}

func getFileContent(r *require.Assertions, filename, extension string) string {
	filename = fmt.Sprintf("%s.%s", filename, extension)
	content, err := ioutil.ReadFile(filename)
	r.Nil(err)

	return string(content)
}
func removeExtension(filename string) string {
	parts := strings.Split(filename, ".")
	return strings.Join(parts[:len(parts)-1], ".")
}

func TestParse(t *testing.T) {
	Suite.SetUpTest(t)
	Suite.TestParse(t)
}

func TestNativeParse(t *testing.T) {
	Suite.SetUpTest(t)
	Suite.TestNativeParse(t)
}
