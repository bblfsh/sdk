package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"gopkg.in/bblfsh/sdk.v1/manifest"
	"gopkg.in/bblfsh/sdk.v1/protocol"

	"google.golang.org/grpc"
)

const FixturesCommandDescription = "" +
	"Generate integration test native and UAST fixtures from source files"

type FixturesCommand struct {
	Args struct {
		SourceFiles []string `positional-arg-name:"sourcefile(s)" required:"true" description:"File(s) with the source code"`
	} `positional-args:"yes"`
	// XXX language autodetect with empty default?
	Language  string `long:"language" required:"true" description:"Language to parse"`
	Endpoint  string `long:"endpoint" default:"localhost:9432" description:"Endpoint of the gRPC server to use"`
	ExtNative string `long:"extnative" default:"native" description:"File extension for native files"`
	ExtUast   string `long:"extuast" default:"uast" description:"File extension for uast files"`

	manifestCommand
	cli protocol.ProtocolServiceClient
}

func (c *FixturesCommand) Execute(args []string) error {
	if err := c.manifestCommand.Execute(args); err != nil {
		return err
	}

	c.processManifest(c.Manifest)
	return nil
}

func (c *FixturesCommand) processManifest(m *manifest.Manifest) {
	conn, err := grpc.Dial(c.Endpoint, grpc.WithTimeout(time.Second*2), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Println("Endpoint connection error, is a bblfshd server running?")
		panic(err)
	}

	c.cli = protocol.NewProtocolServiceClient(conn)

	for _, f := range c.Args.SourceFiles {
		err := c.generateFixtures(f)
		if err != nil {
			fmt.Println("While generating fixtures for ", f)
			panic(err)
		}
	}
}

func (c *FixturesCommand) generateFixtures(f string) error {
	fmt.Println("Processing", f, "...")

	source, err := getSourceFile(f)
	if err != nil {
		return err
	}

	native, err := c.getNative(source)
	if err != nil {
		return err
	}

	err = writeResult(f, native, c.ExtNative)
	if err != nil {
		return err
	}

	uast, err := c.getUast(source)
	if err != nil {
		return err
	}

	err = writeResult(f, uast, c.ExtUast)
	if err != nil {
		return err
	}

	return nil
}

func (c *FixturesCommand) getNative(source string) (string, error) {
	req := &protocol.NativeParseRequest{
		Language: c.Language,
		Content:  source,
	}

	res, err := c.cli.NativeParse(context.Background(), req)
	if err != nil {
		return "", err
	}

	strres, err := NativeParseResponseToString(res)
	if err != nil {
		return "", err
	}

	return strres, nil
}

func (c *FixturesCommand) getUast(source string) (string, error) {
	req := &protocol.ParseRequest{
		Language: c.Language,
		Content:  source,
	}

	res, err := c.cli.Parse(context.Background(), req)
	if err != nil {
		return "", err
	}

	return res.String(), nil
}

func getSourceFile(f string) (string, error) {
	content, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// XXX factorize with the version in sdk/driver/integration/suite_test.go
func NativeParseResponseToString(res *protocol.NativeParseResponse) (string, error) {
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
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	e := json.NewEncoder(buf)
	e.SetIndent("", "    ")
	e.SetEscapeHTML(false)

	err = e.Encode(s)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// XXX factorize
func removeExtension(filename string) string {
	parts := strings.Split(filename, ".")
	return strings.Join(parts[:len(parts)-1], ".")
}

func writeResult(origname, content, extension string) error {
	outname := removeExtension(origname) + "." + extension
	fmt.Println("\tWriting", outname, "...")

	err := ioutil.WriteFile(outname, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}
