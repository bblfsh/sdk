package jsonlines

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

const (
	// DefaultBufferSize is the default buffer size for decoding. It will
	// be used whenever the given reader is not buffered.
	DefaultBufferSize = 1024 * 1024
)

type lineReader interface {
	ReadLine() ([]byte, bool, error)
}

// Decoder decodes JSON lines.
type Decoder interface {
	// Decode decodes the next JSON line into the given value.
	Decode(interface{}) error
}

type decoder struct {
	r lineReader
}

// NewDecoder creates a new decoder with the given reader. If the given reader
// is not buffered, it will be wrapped with a *bufio.Reader.
func NewDecoder(r io.Reader) Decoder {
	var lr lineReader
	if v, ok := r.(lineReader); ok {
		lr = v
	} else {
		lr = bufio.NewReaderSize(r, DefaultBufferSize)
	}

	return &decoder{r: lr}
}

func (d *decoder) Decode(v interface{}) error {
	line, isPrefix, err := d.r.ReadLine()
	if isPrefix {
		return fmt.Errorf("buffer size exceeded")
	}

	if err != nil {
		return err
	}

	return json.Unmarshal(line, v)
}
