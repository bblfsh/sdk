package native

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"

	"github.com/bblfsh/sdk/v3/driver"
	derrors "github.com/bblfsh/sdk/v3/driver/errors"
	"github.com/bblfsh/sdk/v3/driver/native/jsonlines"
	"github.com/bblfsh/sdk/v3/uast/nodes"
	serrors "gopkg.in/src-d/go-errors.v1"
)

var (
	// Binary default location of the native driver binary. Should not
	// override this variable unless you know what are you doing.
	Binary = "/opt/driver/bin/native"
)

const (
	closeTimeout = time.Second * 5
)

var (
	// ErrNotRunning is returned when calling Parse on the not running driver.
	ErrNotRunning = serrors.NewKind("native driver is not running")
	// ErrDriverCrashed is returned when the driver crashes after parsing attempt.
	ErrDriverCrashed = serrors.NewKind("native driver crashed")
)

func NewDriver(enc Encoding) driver.Native {
	return NewDriverAt("", enc)
}

func NewDriverAt(bin string, enc Encoding) driver.Native {
	if bin == "" {
		bin = Binary
	}
	if enc == "" {
		enc = UTF8
	}
	return &Driver{bin: bin, ec: enc}
}

type driverState int

const (
	stateOK = driverState(iota)
	stateTimeout
	stateBroken
)

// Driver is a wrapper of the native command. The operations with the
// driver are synchronous by design, this is controlled by a mutex. This means
// that only one parse request can attend at the same time.
type Driver struct {
	bin     string
	ec      Encoding
	started bool

	mu     sync.Mutex
	enc    jsonlines.Encoder
	dec    jsonlines.Decoder
	stdin  *os.File
	stdout *os.File
	cmd    *exec.Cmd
	cmdErr chan error
	state  driverState
}

// Start executes the given native driver and prepares it to parse code.
func (d *Driver) Start() error {
	d.state = stateOK
	d.cmd = exec.Command(d.bin)
	d.cmd.Stderr = os.Stderr

	var (
		err           error
		stdin, stdout *os.File
	)

	stdin, d.stdin, err = os.Pipe()
	if err != nil {
		return err
	}

	d.stdout, stdout, err = os.Pipe()
	if err != nil {
		stdin.Close()
		d.stdin.Close()
		return err
	}
	d.cmd.Stdin = stdin
	d.cmd.Stdout = stdout

	d.enc = jsonlines.NewEncoder(d.stdin)
	d.dec = jsonlines.NewDecoder(d.stdout)

	err = d.cmd.Start()
	if err == nil {
		d.started = true
		errc := make(chan error, 1)
		d.cmdErr = errc
		go func() {
			// close pipes when driver exits
			defer func() {
				stdin.Close()
				stdout.Close()
				close(errc)
			}()
			errc <- d.cmd.Wait()
		}()
		return nil
	}
	d.stdin.Close()
	d.stdout.Close()
	stdin.Close()
	stdout.Close()
	return err
}

// parseRequest is the request used to communicate the driver with the
// native driver via json.
type parseRequest struct {
	Content  string   `json:"content"`
	Encoding Encoding `json:"Encoding"`
}

var _ json.Unmarshaler = (*parseResponse)(nil)

// parseResponse is the reply to parseRequest by the native parser.
type parseResponse struct {
	Status status     `json:"status"`
	Errors []string   `json:"errors"`
	AST    nodes.Node `json:"ast"`
}

func (r *parseResponse) UnmarshalJSON(data []byte) error {
	var resp struct {
		Status status      `json:"status"`
		Errors []string    `json:"errors"`
		AST    interface{} `json:"ast"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}
	ast, err := nodes.ToNode(resp.AST, nil)
	if err != nil {
		return err
	}
	*r = parseResponse{
		Status: resp.Status,
		Errors: resp.Errors,
		AST:    ast,
	}
	return nil
}

func (d *Driver) writeRequest(ctx context.Context, req *parseRequest) error {
	sp, _ := opentracing.StartSpanFromContext(ctx, "bblfsh.native.Parse.encodeReq")
	defer sp.Finish()

	err := d.enc.Encode(req)
	if err == nil {
		return nil
	}
	// Cannot write data - this means the stream is broken or driver crashed.
	// We will try to recover by reading the response, but since it might be
	// a stack trace or an error message, we will read it as a "raw" value.
	// This preserves an original text instead of failing with decoding error.
	var raw json.RawMessage
	// TODO: this reads a single line only; we can be smarter and read the whole log if driver cannot recover
	if err := d.dec.Decode(&raw); err != nil {
		// stream is broken on both sides, cannot get additional info
		return driver.ErrDriverFailure.Wrap(err)
	}
	return driver.ErrDriverFailure.Wrap(fmt.Errorf("error: %v; %s", err, string(raw)))
}

type timeoutError interface {
	Timeout() bool
}

func (d *Driver) broken() {
	d.state = stateBroken
	_ = d.close()
}

func (d *Driver) skipResponse(ctx context.Context) error {
	if d.state != stateTimeout { // safeguard
		panic(fmt.Errorf("unexpected state: %v", d.state))
	}
	sp, _ := opentracing.StartSpanFromContext(ctx, "bblfsh.native.Parse.skipResp")
	defer sp.Finish()

	// TODO(dennwc): relies on JSON; should probably be a method on an interface
	var r json.RawMessage
	err := d.dec.Decode(&r)
	if e, ok := err.(timeoutError); ok && e.Timeout() {
		d.state = stateTimeout
		return err
	} else if err != nil {
		d.broken()
		return err
	}
	d.state = stateOK
	return nil
}

func (d *Driver) readResponse(ctx context.Context) (*parseResponse, error) {
	sp, _ := opentracing.StartSpanFromContext(ctx, "bblfsh.native.Parse.decodeResp")
	defer sp.Finish()

	var r parseResponse
	err := d.dec.Decode(&r)
	if e, ok := err.(timeoutError); ok && e.Timeout() {
		// the request is still being processed by the native driver,
		// so next time we will need to discard the first response
		d.state = stateTimeout
		return nil, err
	} else if err != nil {
		// we can't be sure what happened, so let's not mess with
		// the client; we will stop the driver now
		d.broken()
		return nil, err
	}
	return &r, nil
}

func (d *Driver) restart() error {
	// driver died; we don't care about exit code
	<-d.cmdErr
	// try to restart it once
	// TODO(dennwc): use exponential backoff? but it may slow down the processing in case
	//				 a user sends a batch of broken files
	//				 maybe we can somehow differentiate between those two cases?
	if err := d.Start(); err != nil {
		return driver.ErrDriverFailure.Wrap(err, "driver restart failed")
	}
	return nil
}

// Parse sends a request to the native driver and returns its response.
func (d *Driver) Parse(rctx context.Context, src string) (nodes.Node, error) {
	sp, ctx := opentracing.StartSpanFromContext(rctx, "bblfsh.native.Parse")
	defer sp.Finish()

	if !d.started {
		return nil, driver.ErrDriverFailure.Wrap(ErrNotRunning.New())
	}

	str, err := d.ec.Encode(src)
	if err != nil {
		return nil, driver.ErrDriverFailure.Wrap(err)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if deadline, ok := ctx.Deadline(); ok {
		_ = d.stdout.SetReadDeadline(deadline)
		_ = d.stdin.SetWriteDeadline(deadline)
		defer func() {
			_ = d.stdin.SetWriteDeadline(time.Time{})
			_ = d.stdout.SetReadDeadline(time.Time{})
		}()
	}

	if d.state == stateTimeout {
		// timed out last time, so we still have a response on the wire
		// skip it before sending a new request
		if err = d.skipResponse(ctx); err != nil {
			return nil, driver.ErrDriverFailure.Wrap(err)
		}
	} else if d.state == stateBroken {
		// protocol is broken and we decided to shutdown the driver
		// try restarting it now
		if err := d.restart(); err != nil {
			return nil, driver.ErrDriverFailure.Wrap(err, "driver restart failed")
		}
	} else if d.state != stateOK {
		return nil, driver.ErrDriverFailure.Wrap(err, "unexpected state: %v", d.state)
	}

	err = d.writeRequest(ctx, &parseRequest{
		Content: str, Encoding: d.ec,
	})
	if err != nil {
		return nil, err
	}

	r, err := d.readResponse(ctx)
	if err == io.EOF {
		if err := d.restart(); err != nil {
			return nil, err
		}
		// fail anyway - this request may have caused the crash
		err = ErrDriverCrashed.New()
	}
	if err != nil {
		return nil, driver.ErrDriverFailure.Wrap(err)
	}
	if r.Status == statusOK {
		return r.AST, nil
	}
	errs := make([]error, 0, len(r.Errors))
	for _, s := range r.Errors {
		errs = append(errs, errors.New(s))
	}
	err = derrors.Join(errs)
	switch r.Status {
	case statusError:
		// parsing error, wrapping will be done on a higher level
	case statusFatal:
		err = driver.ErrDriverFailure.Wrap(err)
		r.AST = nil // do not allow to propagate AST with Fatal error
	default:
		return nil, fmt.Errorf("unsupported status: %v", r.Status)
	}
	return r.AST, err
}

// close stops the execution of the native driver.
func (d *Driver) close() error {
	// note: it should not hold the mutex, or readResponse will deadlock
	var last error
	if err := d.stdin.Close(); err != nil {
		last = err
	}
	if er, ok := last.(*os.PathError); ok && er.Err == os.ErrClosed {
		last = nil
	}
	timeout := time.NewTimer(closeTimeout)
	select {
	case <-d.cmdErr: // don't care about exit code
		timeout.Stop()
	case <-timeout.C:
		d.cmd.Process.Kill()
	}
	err := d.stdout.Close()
	if last != nil {
		return last
	}
	if er, ok := err.(*os.PathError); ok && er.Err == os.ErrClosed {
		err = nil
	}
	if err != nil {
		last = err
	}
	return last
}

// Close stops the execution of the native driver.
func (d *Driver) Close() error {
	if !d.started {
		return nil
	}
	d.started = false
	return d.close()
}

var _ json.Unmarshaler = (*status)(nil)

type status string

func (s *status) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	str = strings.ToLower(str)
	*s = status(str)
	return nil
}

const (
	statusOK = status("ok")
	// statusError is replied when the driver has got the AST with errors.
	statusError = status("error")
	// statusFatal is replied when the driver hasn't could get the AST.
	statusFatal = status("fatal")
)

var _ json.Unmarshaler = (*Encoding)(nil)

// Encoding is the Encoding used for the content string. Currently only
// UTF-8 or Base64 encodings are supported. You should use UTF-8 if you can
// and Base64 as a fallback.
type Encoding string

const (
	UTF8   = Encoding("utf8")
	Base64 = Encoding("base64")
)

func (e *Encoding) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	str = strings.ToLower(str)
	*e = Encoding(str)
	return nil
}

// Encode converts UTF8 string into specified Encoding.
func (e Encoding) Encode(s string) (string, error) {
	switch e {
	case UTF8:
		return s, nil
	case Base64:
		s = base64.StdEncoding.EncodeToString([]byte(s))
		return s, nil
	default:
		return "", fmt.Errorf("invalid Encoding: %v", e)
	}
}

// Decode converts specified Encoding into UTF8.
func (e Encoding) Decode(s string) (string, error) {
	switch e {
	case UTF8:
		return s, nil
	case Base64:
		b, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return "", err
		}
		return string(b), nil
	default:
		return "", fmt.Errorf("invalid Encoding: %v", e)
	}
}
