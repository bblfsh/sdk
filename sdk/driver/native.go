package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"gopkg.in/bblfsh/sdk.v1/protocol"
	"gopkg.in/bblfsh/sdk.v1/sdk/jsonlines"

	"srcd.works/go-errors.v0"
)

var (
	// NativeBinary default location of the native driver binary. Should not
	// override this variable unless you know what are you doing.
	NativeBinary = "/opt/driver/bin/native"

	ErrUnsupportedLanguage = errors.NewKind("unsupported language got %q, expected %q")
)

// NativeDriver is a wrapper of the native command. The operations with the
// driver are synchronous by design, this is controlled by a mutex. This means
// that only one parse request can attend at the same time.
type NativeDriver struct {
	enc    jsonlines.Encoder
	dec    jsonlines.Decoder
	closer io.Closer
	cmd    *exec.Cmd

	m sync.Mutex
}

// Start executes the given native driver and prepares it to parse code.
func (d *NativeDriver) Start() error {
	d.cmd = exec.Command(NativeBinary)
	stdin, err := d.cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := d.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	d.enc = jsonlines.NewEncoder(stdin)
	d.dec = jsonlines.NewDecoder(stdout)
	d.closer = stdin

	stderr, err := d.cmd.StderrPipe()
	if err != nil {
		return err
	}

	go io.Copy(os.Stderr, stderr)

	return d.cmd.Start()
}

// Parse sends a request to the native driver and returns its response.
func (d *NativeDriver) Parse(req *InternalParseRequest) *InternalParseResponse {
	d.m.Lock()
	defer d.m.Unlock()

	_ = d.enc.Encode(&InternalParseRequest{
		Content:  req.Content,
		Encoding: Encoding(req.Encoding),
	})

	r := &InternalParseResponse{}
	if err := d.dec.Decode(r); err != nil {
		r.Status = Status(protocol.Fatal)
		r.Errors = append(r.Errors, err.Error())
	}

	return r
}

// Stop stops the execution of the native driver.
func (d *NativeDriver) Stop() error {
	if err := d.closer.Close(); err != nil {
		return err
	}

	return d.cmd.Wait()
}

// InternalParseRequest is the request used to communicate the driver with the
// native driver via json.
type InternalParseRequest struct {
	Content  string   `json:"content"`
	Encoding Encoding `json:"encoding"`
}

// InternalParseResponse is the reply to InternalParseRequest by the native
// parser.
type InternalParseResponse struct {
	Status Status      `json:"status"`
	Errors []string    `json:"errors"`
	AST    interface{} `json:"ast"`
}

type Status protocol.Status

func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(protocol.Status(s).String())
}

func (s *Status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Status should be a string, got %s", data)
	}

	i, ok := protocol.Status_value[strings.ToUpper(str)]
	if !ok {
		return fmt.Errorf("Unknown status got %q", str)
	}

	*s = Status(i)
	return nil
}

type Encoding protocol.Encoding

func (e Encoding) MarshalJSON() ([]byte, error) {
	return json.Marshal(protocol.Encoding(e).String())
}

func (e *Encoding) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("Encoding should be a string, got %s", data)
	}

	i, ok := protocol.Encoding_value[strings.ToUpper(str)]
	if !ok {
		return fmt.Errorf("Unknown status got %q", str)
	}

	*e = Encoding(i)
	return nil
}