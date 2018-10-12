package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"google.golang.org/grpc"
	protocol1 "gopkg.in/bblfsh/sdk.v1/protocol"
	protocol2 "gopkg.in/bblfsh/sdk.v2/protocol"

	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/bblfsh/sdk.v2/uast/yaml"
)

const FixturesCommandDescription = "" +
	"Generate integration tests' '.native' and '.uast' fixtures from source files"

type FixturesCommand struct {
	Args struct {
		SourceFiles []string `positional-arg-name:"sourcefile(s)" required:"true" description:"File(s) with the source code"`
	} `positional-args:"yes"`
	Language  string `long:"language" short:"l" default:"" description:"Language to parse"`
	Endpoint  string `long:"endpoint" short:"e" default:"localhost:9432" description:"Endpoint of the gRPC server to use"`
	ExtNative string `long:"extnative" short:"n" default:"native" description:"File extension for native files"`
	ExtUast   string `long:"extuast" short:"u" default:"uast" description:"File extension for uast files"`
	ExtProto  string `long:"extproto" short:"p" description:"File extenstion for proto message files"`
	Quiet     bool   `long:"quiet" short:"q" description:"Don't print any output"`

	cli1 protocol1.ProtocolServiceClient
	cli2 protocol2.DriverClient
}

func (c *FixturesCommand) Execute(args []string) error {
	conn, err := grpc.Dial(c.Endpoint, grpc.WithTimeout(time.Second*2), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Println("Endpoint connection error, is a bblfshd server running?")
		return err
	}

	c.cli1 = protocol1.NewProtocolServiceClient(conn)
	c.cli2 = protocol2.NewDriverClient(conn)

	for _, f := range c.Args.SourceFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			fmt.Println("Error: File", f, "doesn't exists")
			os.Exit(1)
		}

		err := c.generateFixtures(f)
		if err != nil {
			fmt.Println("While generating fixtures for ", f)
			return err
		}
	}
	return nil
}

//generateFixtures writes .uast, .native and .proto (optional) files.
//All of them contain plain-text represenation of the same UAST in differetn formats:
//v1, v2 native, v1 proto.

func (c *FixturesCommand) generateFixtures(filename string) error {
	if !c.Quiet {
		fmt.Println("Processing", filename, "...")
	}

	source, err := getSourceFile(filename)
	if err != nil {
		return err
	}

	err = c.writeNative(source, filename, c.ExtNative)
	if err != nil {
		return err
	}

	//TODO(bzz): refactor/add writeSemm(), writeLegacy(), writeProto()

	uast, err := c.getUast(source, filename)
	if err != nil {
		return err
	}

	err = c.writeResult(filename, c.ExtUast, []byte(uast.String()))
	if err != nil {
		return err
	}

	if c.ExtProto != "" {
		protoUast, err := uast.UAST.Marshal()
		if err != nil {
			return err
		}
		err = c.writeResult(filename, c.ExtProto, protoUast)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *FixturesCommand) writeNative(source, filename, ext string) error {
	ast, err := c.getNative(source, filename)
	if err != nil {
		return err
	}

	data, err := uastyml.Marshal(ast)
	if err != nil {
		return err
	}

	return c.writeResult(filename, ext, data)
}

func (c *FixturesCommand) getNative(source string, filename string) (nodes.Node, error) {
	req := &protocol2.ParseRequest{
		Language: c.Language,
		Content:  source,
		Filename: filename,
		Mode:     protocol2.Mode_Native,
	}

	res, err := c.cli2.Parse(context.Background(), req)
	if err != nil {
		return nil, err
	}

	ast, err := res.Nodes()
	if err != nil {
		if !c.Quiet {
			fmt.Println("Warning: parsing native AST for ", filename, "returned errors:")
			fmt.Println(err)
		}
	}

	return ast, nil
}

func (c *FixturesCommand) getUast(source string, filename string) (*protocol1.ParseResponse, error) {
	req := &protocol1.ParseRequest{
		Language: c.Language,
		Content:  source,
		Filename: filename,
	}

	res, err := c.cli1.Parse(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if res.Status != protocol1.Ok {
		if !c.Quiet {
			fmt.Println("Warning: parse request for ", filename, "returned errors:")
			for _, e := range res.Errors {
				fmt.Println(e)
			}
		}
	}

	return res, nil
}

func (c *FixturesCommand) writeResult(origName, extension string, content []byte) error {
	outname := origName + "." + extension
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
