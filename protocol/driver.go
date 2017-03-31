package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/bblfsh/sdk"
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
	// ToNoder converts original AST to *uast.Node.
	ToNoder uast.ToNoder
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
	if err := d.run(os.Args); err != nil {
		os.Exit(1)
	}
}

func (d *Driver) run(args []string) error {
	d.initialize()

	cmd := cmd{Driver: d}
	cmd.NativeBin = sdk.NativeBin

	if len(os.Args) == 1 {
		cmd := &serveCommand{cmd}
		return cmd.Execute(nil)
	}

	parser := flags.NewNamedParser(args[0], flags.HelpFlag)
	parser.AddCommand("serve", "", "", &serveCommand{cmd: cmd})
	parser.AddCommand("parse-native", "", "", &parseNativeASTCommand{cmd: cmd})
	parser.AddCommand("parse-uast", "", "", &parseUASTCommand{cmd: cmd})

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
	NativeBin string `long:"native-bin" description:"alternative path for the native binary"`
}

type serveCommand struct {
	cmd
}

func (c *serveCommand) Execute(args []string) error {
	client, err := ExecNative(c.NativeBin)
	if err != nil {
		return fmt.Errorf("error executing native: %s", err.Error())
	}

	server := &Server{
		In:       c.In,
		Out:      c.Out,
		Native:   client,
		ToNoder:  c.ToNoder,
		Annotate: c.Annotate,
	}

	if err := server.Start(); err != nil {
		_ = client.Close()
		return fmt.Errorf("error starting server: %s", err.Error())
	}

	if err := server.Wait(); err != nil {
		_ = client.Close()
		return fmt.Errorf("error waiting for server end: %s", err.Error())
	}

	if err := client.Close(); err != nil {
		return fmt.Errorf("error closing native: %s", err.Error())
	}

	return nil
}

type parseNativeASTCommand struct {
	cmd
	Args struct {
		File string
	} `positional-args:"yes"`
}

func (c *parseNativeASTCommand) Execute(args []string) error {
	f := c.Args.File

	client, err := ExecNative(c.NativeBin)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("error reading file %s: %s", f, err.Error())
	}

	req := &ParseNativeASTRequest{
		Content: string(b),
	}

	resp, err := client.ParseNativeAST(req)
	if err != nil {
		return fmt.Errorf("request failed: %q", err)
	}

	e := json.NewEncoder(c.Out)
	if err := e.Encode(resp); err != nil {
		return err
	}

	return nil
}

type parseUASTCommand struct {
	cmd
	Format string `long:"format" default:"json" description:"json, pretty (default: json)"`
	Args   struct {
		File string
	} `positional-args:"yes"`
}

func (c *parseUASTCommand) Execute(args []string) error {
	f := c.Args.File

	var formatter func(io.Writer, *ParseUASTResponse) error
	switch c.Format {
	case "pretty":
		formatter = c.prettyPrinter
	case "json":
		formatter = c.jsonPrinter
	default:
		return fmt.Errorf("invalid format: %s", c.Format)
	}

	client, err := ExecNative(c.NativeBin)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("error reading file %s: %s", f, err.Error())
	}

	req := &ParseNativeASTRequest{
		Content: string(b),
	}

	resp, err := client.ParseNativeAST(req)
	if err != nil {
		return fmt.Errorf("request failed: %q", err)
	}

	uastResp := &ParseUASTResponse{}
	uastResp.Status = resp.Status
	uastResp.Errors = resp.Errors
	if err == nil && resp.Status != Fatal {
		n, err := c.ToNoder.ToNode(resp.AST)
		if err != nil {
			uastResp.Status = Fatal
			uastResp.Errors = append(uastResp.Errors, err.Error())
		} else {
			if err := c.Annotate.Apply(n); err != nil {
				uastResp.Status = Error
				uastResp.Errors = append(uastResp.Errors, err.Error())
			}

			uastResp.UAST = n
		}
	}

	return formatter(c.Out, uastResp)
}

func (c *parseUASTCommand) jsonPrinter(w io.Writer, r *ParseUASTResponse) error {
	e := json.NewEncoder(w)
	return e.Encode(r)
}

func (c *parseUASTCommand) prettyPrinter(w io.Writer, r *ParseUASTResponse) error {
	fmt.Fprintln(w, "Status: ", r.Status)
	fmt.Fprintln(w, "Errors: ")
	for _, err := range r.Errors {
		fmt.Fprintln(w, " . ", err)
	}
	fmt.Fprintln(w, "UAST: ")
	fmt.Fprintln(w, r.UAST.String())
	return nil
}
