package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"gopkg.in/bblfsh/sdk.v2/protocol"
	"gopkg.in/bblfsh/sdk.v2/sdk/jsonlines"

	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/bblfsh/sdk.v2/uast/nodes"
	"gopkg.in/src-d/go-errors.v1"
)

var (
	// NativeBinary default location of the native driver binary. Should not
	// override this variable unless you know what are you doing.
	NativeBinary = "/opt/driver/bin/native"
)

var (
	ErrUnsupportedLanguage = errors.NewKind("unsupported language got %q, expected %q")
	ErrNativeNotRunning    = errors.NewKind("native driver is not running")
)

// NativeMain is a main function for running a native Go driver as an Exec-based module that uses internal json protocol.
func NativeMain(d BaseDriver) {
	if err := d.Start(); err != nil {
		panic(err)
	}
	defer d.Close()
	srv := &nativeServer{d: d}
	c := struct {
		io.Reader
		io.Writer
	}{
		os.Stdin,
		os.Stdout,
	}
	if err := srv.Serve(c); err != nil {
		panic(err)
	}
}

type nativeServer struct {
	d BaseDriver
}

func errResp(err error) *InternalParseResponse {
	return &InternalParseResponse{
		Status: Status(protocol.Fatal),
		Errors: []string{err.Error()},
	}
}

func errRespf(format string, args ...interface{}) *InternalParseResponse {
	return errResp(fmt.Errorf(format, args...))
}

func (s *nativeServer) Serve(c io.ReadWriter) error {
	enc := jsonlines.NewEncoder(c)
	dec := jsonlines.NewDecoder(c)
	for {
		var req InternalParseRequest
		err := dec.Decode(&req)
		if err == io.EOF {
			return nil
		} else if err != nil {
			if err = enc.Encode(errRespf("failed to decode request: %v", err)); err != nil {
				return err
			}
			continue
		}
		resp, err := s.d.Parse(&req)
		if err != nil {
			if err := enc.Encode(errResp(err)); err != nil {
				return err
			}
			continue
		}
		if err = enc.Encode(resp); err != nil {
			return err
		}
	}
}

func NewExecDriver() BaseDriver {
	return NewExecDriverAt("")
}

func NewExecDriverAt(bin string) BaseDriver {
	if bin == "" {
		bin = NativeBinary
	}
	return &ExecDriver{bin: bin}
}

// ExecDriver is a wrapper of the native command. The operations with the
// driver are synchronous by design, this is controlled by a mutex. This means
// that only one parse request can attend at the same time.
type ExecDriver struct {
	bin     string
	running bool

	mu     sync.Mutex
	enc    jsonlines.Encoder
	dec    jsonlines.Decoder
	stdin  io.Closer
	stdout io.Closer
	cmd    *exec.Cmd
}

// Start executes the given native driver and prepares it to parse code.
func (d *ExecDriver) Start() error {
	d.cmd = exec.Command(d.bin)
	d.cmd.Stderr = os.Stderr

	stdin, err := d.cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err := d.cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return err
	}

	d.stdin = stdin
	d.stdout = stdout
	d.enc = jsonlines.NewEncoder(stdin)
	d.dec = jsonlines.NewDecoder(stdout)

	err = d.cmd.Start()
	if err == nil {
		d.running = true
		return nil
	}
	stdin.Close()
	stdout.Close()
	return err
}

// Parse sends a request to the native driver and returns its response.
func (d *ExecDriver) Parse(req *InternalParseRequest) (*InternalParseResponse, error) {
	if !d.running {
		return nil, ErrNativeNotRunning.New()
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	_ = d.enc.Encode(&InternalParseRequest{
		Content:  req.Content,
		Encoding: Encoding(req.Encoding),
	})

	r := &InternalParseResponse{}
	if err := d.dec.Decode(r); err != nil {
		r.Status = Status(protocol.Fatal)
		r.Errors = append(r.Errors, err.Error())
	}

	return r, nil
}

// Stop stops the execution of the native driver.
func (d *ExecDriver) Close() error {
	var last error
	if err := d.stdin.Close(); err != nil {
		last = err
	}
	err := d.cmd.Wait()
	err2 := d.stdout.Close()
	if err != nil {
		return err
	}
	if er, ok := err2.(*os.PathError); ok && er.Err == os.ErrClosed {
		err2 = nil
	}
	if err2 != nil {
		last = err2
	}
	return last
}

// InternalParseRequest is the request used to communicate the driver with the
// native driver via json.
type InternalParseRequest struct {
	Content  string   `json:"content"`
	Encoding Encoding `json:"encoding"`
}

var _ json.Unmarshaler = (*InternalParseResponse)(nil)

// InternalParseResponse is the reply to InternalParseRequest by the native
// parser.
type InternalParseResponse struct {
	Status Status     `json:"status"`
	Errors []string   `json:"errors"`
	AST    nodes.Node `json:"ast"`
}

func (r *InternalParseResponse) UnmarshalJSON(data []byte) error {
	var resp struct {
		Status Status      `json:"status"`
		Errors []string    `json:"errors"`
		AST    interface{} `json:"ast"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	ast, err := uast.ToNode(resp.AST)
	if err != nil {
		return err
	}
	*r = InternalParseResponse{
		Status: resp.Status,
		Errors: resp.Errors,
		AST:    ast,
	}
	return nil
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
