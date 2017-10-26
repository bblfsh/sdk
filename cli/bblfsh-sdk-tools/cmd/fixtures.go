package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"
	"os"

	common "gopkg.in/bblfsh/sdk.v1"
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
	Language  string `long:"language" required:"true" description:"Language to parse"`
	Endpoint  string `long:"endpoint" default:"localhost:9432" description:"Endpoint of the gRPC server to use"`
	ExtNative string `long:"extnative" default:"native" description:"File extension for native files"`
	ExtUast   string `long:"extuast" default:"uast" description:"File extension for uast files"`
	Quiet     bool   `long:"quiet" description:"Don't print any output"`

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
		if _, err := os.Stat(f); os.IsNotExist(err) {
			// path/to/whatever does not exist
			fmt.Println("Error: File", f, "doesn't exists")
			os.Exit(1)
		}

		err := c.generateFixtures(f)
		if err != nil {
			fmt.Println("While generating fixtures for ", f)
			panic(err)
		}
	}
}

func (c *FixturesCommand) generateFixtures(f string) error {
	if !c.Quiet {
		fmt.Println("Processing", f, "...")
	}

	source, err := getSourceFile(f)
	if err != nil {
		return err
	}

	native, err := c.getNative(source)
	if err != nil {
		return err
	}

	err = c.writeResult(f, native, c.ExtNative)
	if err != nil {
		return err
	}

	uast, err := c.getUast(source)
	if err != nil {
		return err
	}

	err = c.writeResult(f, uast, c.ExtUast)
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

	strres, err := common.NativeParseResponseToString(res)
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

func (c *FixturesCommand) writeResult(origname, content, extension string) error {
	outname := common.RemoveExtension(origname) + "." + extension
	if !c.Quiet {
		fmt.Println("\tWriting", outname, "...")
	}

	err := ioutil.WriteFile(outname, []byte(content), 0644)
	if err != nil {
		return err
	}

	return nil
}

func getSourceFile(f string) (string, error) {
	content, err := ioutil.ReadFile(f)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

