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
	DefaultBufferSize = 1024 * 1024 * 4
)

type lineReader interface {
	ReadLine() ([]byte, bool, error)
	ReadSlice(delim byte) (line []byte, err error)
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

// Read and discard everything until the next newline
func (d *decoder) discardPending() error {
	for {
		_, err := d.r.ReadSlice('\n')
		switch err {
		case io.EOF:
			return nil
		case bufio.ErrBufferFull:
			continue
		case nil:
			continue
		default:
			return err
		}
	}
}

func (d *decoder) readLine() (line []byte, err error) {
	for {
		chunk, isPrefix, err := d.r.ReadLine()
		if err != nil {
			if !isPrefix {
				// Discard pending contents in the input buffer to avoid IO blocking
				discardErr := d.discardPending()
				if discardErr != nil {
					err = fmt.Errorf("%s; aditionally, there was an error "+
						"trying to dicard the input buffer: %s",
						err, discardErr)
				}
			}
			return nil, err
		}
		line = append(line, chunk...)

		if !isPrefix {
			// EOL found.
			break
		}
	}

	return line, nil
}

// Decode decodes the next line in the reader.
// It does not check JSON for well-formedness before decoding, so in case of
// error, the structure might be half-filled.
func (d *decoder) Decode(v interface{}) error {
	line, err := d.readLine()
	if err != nil {
		return err
	}

	switch o := v.(type) {
	case json.Unmarshaler:
		return o.UnmarshalJSON(line)
	default:
		return json.Unmarshal(line, v)
	}
}
