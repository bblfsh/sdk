package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bblfsh/sdk/protocol"
	"github.com/bblfsh/sdk/protocol/native"
	"github.com/bblfsh/sdk/uast"
	"github.com/bblfsh/sdk/uast/ann"

	"github.com/jessevdk/go-flags"
)

// Driver implements a driver.
type Driver struct {
	// Version of the driver.
	Version string
	// Build identifier.
	Build string
	// ASTParserBuilder creates a ASTParser.
	ASTParserBuilder ASTParserBuilder
	// Annotate contains an *ann.Rule to convert AST to UAST.
	Annotate *ann.Rule
	// In is the input of the driver. Defaults to os.Stdin.
	In io.Reader
	// Out is the output of the driver. Defaults to os.Stdout.
	Out io.Writer
	// Err is the error output of the driver. Defaults to os.Stderr.
	Err io.Writer
}

// Exec runs the driver with the given command line arguments.
// Note that this method contains calls to os.Exit.
func (d *Driver) Exec() {
	if err := d.Run(os.Args); err != nil {
		os.Exit(1)
	}
}

func (d *Driver) Run(args []string) error {
	d.initialize()

	cmd := cmd{Driver: d}

	if len(args) == 1 {
		cmd := &serveCommand{cmd}
		return cmd.Execute(nil)
	}

	parser := flags.NewNamedParser(args[0], flags.HelpFlag)
	parser.AddCommand("serve", "", "", &serveCommand{cmd: cmd})
	parser.AddCommand("parse-native", "", "", &parseNativeASTCommand{cmd: cmd})
	parser.AddCommand("parse-uast", "", "", &parseUASTCommand{cmd: cmd})
	parser.AddCommand("tokenize", "", "", &tokenizeCommand{cmd: cmd})

	if _, err := parser.ParseArgs(args[1:]); err != nil {
		if err, ok := err.(*flags.Error); ok {
			parser.WriteHelp(d.Out)
			fmt.Fprintf(d.Out, "\nBuild information\n  commit: %s\n  date:%s\n", d.Version, d.Build)
			if err.Type == flags.ErrHelp {
				return nil
			}
		}

		return err
	}

	return nil
}

func (d *Driver) initialize() {
	if d.In == nil {
		d.In = os.Stdin
	}

	if d.Out == nil {
		d.Out = os.Stdout
	}

	if d.Err == nil {
		d.Err = os.Stderr
	}
}

type cmd struct {
	*Driver
	ASTParserOptions
}

type serveCommand struct {
	cmd
}

func (c *serveCommand) Execute(args []string) error {
	p, err := c.ASTParserBuilder(c.ASTParserOptions)
	if err != nil {
		return err
	}

	server := &Server{
		In:  c.In,
		Out: c.Out,
		UASTParser: &uastParser{
			ASTParser:  p,
			Annotation: c.Driver.Annotate,
		},
	}

	if err := server.Start(); err != nil {
		_ = p.Close()
		return fmt.Errorf("error starting server: %s", err.Error())
	}

	if err := server.Wait(); err != nil {
		_ = p.Close()
		return fmt.Errorf("error waiting for server end: %s", err.Error())
	}

	if err := p.Close(); err != nil {
		return fmt.Errorf("error closing parser: %s", err.Error())
	}

	return nil
}

type parseNativeASTCommand struct {
	cmd
	Format string `long:"format" default:"json" description:"json, prettyjson (default: json)"`
	Args   struct {
		File string
	} `positional-args:"yes"`
}

func (c *parseNativeASTCommand) Execute(args []string) error {
	f := c.Args.File

	b, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("error reading file %s: %s", f, err.Error())
	}

	nc, err := native.ExecClient(c.NativeBin)
	if err != nil {
		return fmt.Errorf("error executing native client: %s", c)
	}

	defer func() { _ = nc.Close() }()

	resp, err := nc.ParseNativeAST(&native.ParseASTRequest{Content: string(b)})
	if err != nil {
		return err
	}

	e := json.NewEncoder(c.Out)
	if c.Format == "prettyjson" {
		e.SetIndent("", "    ")
	}

	if err := e.Encode(resp); err != nil {
		return err
	}

	return nil
}

type parseUASTCommand struct {
	cmd
	Format string `long:"format" default:"json" description:"json, prettyjson, pretty (default: json)"`
	Args   struct {
		File string
	} `positional-args:"yes"`
}

func (c *parseUASTCommand) Execute(args []string) error {
	fmter, err := formatter(c.Format)
	if err != nil {
		return err
	}

	f := c.Args.File
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("error reading file %s: %s", f, err.Error())
	}

	p, err := c.ASTParserBuilder(c.ASTParserOptions)
	if err != nil {
		return err
	}

	defer func() { _ = p.Close() }()

	up := &uastParser{
		ASTParser:  p,
		Annotation: c.Driver.Annotate,
	}

	resp, err := up.ParseUAST(&protocol.ParseUASTRequest{Content: string(b)})
	if err != nil {
		return err
	}

	return fmter(c.Out, resp)
}

func formatter(f string) (func(io.Writer, *protocol.ParseUASTResponse) error, error) {
	switch f {
	case "pretty":
		return prettyPrinter, nil
	case "prettyjson":
		return prettyJsonPrinter, nil
	case "json":
		return jsonPrinter, nil
	default:
		return nil, fmt.Errorf("invalid format: %s", f)
	}
}

func jsonPrinter(w io.Writer, r *protocol.ParseUASTResponse) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func prettyJsonPrinter(w io.Writer, r *protocol.ParseUASTResponse) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "    ")
	return e.Encode(r)
}

func prettyPrinter(w io.Writer, r *protocol.ParseUASTResponse) error {
	fmt.Fprintln(w, "Status: ", r.Status)
	fmt.Fprintln(w, "Errors: ")
	for _, err := range r.Errors {
		fmt.Fprintln(w, " . ", err)
	}
	fmt.Fprintln(w, "UAST: ")
	fmt.Fprintln(w, r.UAST.String())
	return nil
}

type tokenizeCommand struct {
	cmd
	Args struct {
		File string
	} `positional-args:"yes"`
}

func (c *tokenizeCommand) Execute(args []string) error {
	f := c.Args.File
	b, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("error reading file %s: %s", f, err.Error())
	}

	p, err := c.ASTParserBuilder(c.ASTParserOptions)
	if err != nil {
		return err
	}

	defer func() { _ = p.Close() }()

	up := &uastParser{
		ASTParser:  p,
		Annotation: c.Driver.Annotate,
	}

	resp, err := up.ParseUAST(&protocol.ParseUASTRequest{Content: string(b)})
	if err != nil {
		return err
	}

	toks := uast.Tokens(resp.UAST)
	_, err = fmt.Fprintf(c.Out, strings.Join(toks, "\t"))
	return err
}
