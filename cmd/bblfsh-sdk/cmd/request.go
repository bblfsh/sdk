package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/bblfsh/sdk/v3/cmd"
)

const RequestCommandDescription = "returns parse request payload"

type RequestCommand struct {
	Input  string `short:"f" long:"file" description:"input source file"`
	Output string `short:"o" long:"output" description:"output json payload"`
	cmd.Command
}

type ParseRequest struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func (r *RequestCommand) Execute(args []string) error {
	if len(r.Input) == 0 {
		return fmt.Errorf("no input source file")
	}

	src, err := ioutil.ReadFile(r.Input)
	if err != nil {
		return err
	}

	var w io.Writer
	if len(r.Output) == 0 {
		w = os.Stdout
	} else {
		w, err = os.Create(r.Output)
		if err != nil {
			return err
		}
	}

	err = json.NewEncoder(w).Encode(ParseRequest{
		Content:  string(src),
		Encoding: "UTF-8",
	})

	if c, ok := w.(io.Closer); ok {
		e := c.Close()
		if err == nil {
			err = e
		}
	}

	return err
}
