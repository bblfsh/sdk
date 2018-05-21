package uastyml

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/yaml.v2"
)

func Marshal(n uast.Node) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := NewEncoder(buf)
	if err := enc.Encode(n); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

const (
	tab  = "   "
	null = "~"
)

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: bufio.NewWriter(w)}
}

type Encoder struct {
	w   *bufio.Writer
	err error
}

func (enc *Encoder) Encode(n uast.Node) error {
	enc.marshal(nil, n, false)
	if enc.err != nil {
		return enc.err
	}
	return enc.w.Flush()
}

func (enc *Encoder) marshal(tabs []byte, n uast.Node, field bool) {
	switch n := n.(type) {
	case nil:
		enc.writeString(null)
	case uast.Object:
		enc.writeObject(tabs, n)
	case uast.Array:
		enc.writeArray(tabs, n)
	case uast.Value:
		enc.writeValue(n, field)
	default:
		enc.err = fmt.Errorf("unexpected type: %T", n)
	}
}

func (enc *Encoder) writeObject(tabs []byte, m uast.Object) {
	if len(m) == 0 {
		enc.writeString("{}")
		return
	}
	enc.writeString("{")
	written := make(map[string]struct{})

	ntabs := append(tabs, []byte(tab)...)
	writeSysKey := func(s string, tab bool) {
		if tab {
			enc.writeString("\n")
			enc.write(ntabs)
		}
		written[s] = struct{}{}
		enc.marshalString(s, false)
		enc.writeString(": ")
	}

	typ := ""
	if v, ok := m[uast.KeyType].(uast.String); ok {
		enc.writeString(" ")
		writeSysKey(uast.KeyType, false)
		enc.marshalString(string(v), true)
		enc.writeString(",")
		typ = string(v)
	}
	if v, ok := m[uast.KeyToken].(uast.Value); ok {
		writeSysKey(uast.KeyToken, true)
		enc.writeValue(v, true)
		enc.writeString(",")
	}
	if v := m.Roles(); len(v) != 0 {
		writeSysKey(uast.KeyRoles, true)
		sort.Slice(v, func(i, j int) bool {
			return v[i].String() < v[j].String()
		})
		enc.writeArray(ntabs, uast.RoleList(v...))
		enc.writeString(",")
	}
	// enforce specific sorting for known types
	emitObj := func(key string) {
		if v, ok := m[key].(uast.Object); ok {
			writeSysKey(key, true)
			enc.marshal(ntabs, v, true)
			enc.writeString(",")
		}
	}
	emitInt := func(key string) {
		if v, ok := m[key].(uast.Int); ok {
			writeSysKey(key, true)
			enc.marshal(ntabs, v, true)
			enc.writeString(",")
		}
	}
	switch typ {
	case uast.TypePositions:
		emitObj(uast.KeyStart)
		emitObj(uast.KeyEnd)
	case uast.TypePosition:
		emitInt(uast.KeyPosOff)
		emitInt(uast.KeyPosLine)
		emitInt(uast.KeyPosCol)
	default:
		emitObj(uast.KeyPos)
	}
	if len(m) != len(written) {
		for _, k := range m.Keys() {
			if _, ok := written[k]; ok {
				continue
			}
			v := m[k]
			enc.writeString("\n")
			enc.write(ntabs)
			enc.marshalString(k, false)
			enc.writeString(": ")
			enc.marshal(ntabs, v, true)
			enc.writeString(",")
		}
	}
	enc.writeString("\n")
	enc.write(tabs)
	enc.writeString("}")
}

func (enc *Encoder) writeArray(tabs []byte, m uast.Array) {
	if len(m) == 0 {
		enc.writeString("[]")
		return
	}
	small := true
	for _, o := range m {
		if _, ok := o.(uast.Value); !ok {
			small = false
			break
		}
	}
	enc.writeString("[")
	ntabs := append(tabs, []byte(tab)...)
	if !small {
		enc.writeString("\n")
		enc.write(tabs)
	}
	for i, o := range m {
		if small {
			if i != 0 {
				enc.writeString(", ")
			}
		} else {
			enc.writeString(tab)
		}
		enc.marshal(ntabs, o, false)
		if !small {
			enc.writeString(",\n")
			enc.write(tabs)
		}
	}
	enc.writeString("]")
}

func (enc *Encoder) writeValue(v uast.Value, field bool) {
	switch v := v.(type) {
	case nil:
		enc.writeString(null)
	case uast.String:
		enc.marshalString(string(v), field)
	case uast.Bool:
		if v {
			enc.writeString("true")
		} else {
			enc.writeString("false")
		}
	case uast.Int:
		enc.writeInt(int64(v))
	case uast.Float:
		enc.writeFloat(float64(v))
	default:
		enc.err = fmt.Errorf("unexpected type: %T", v)
	}
}
func (enc *Encoder) marshalString(s string, field bool) {
	var kind stringFormat
	if field {
		kind = stringDoubleQuoted
	} else {
		kind = bestStringFormat(s)
	}
	switch kind {
	case stringPlain:
		// do nothing
	case stringQuoted:
		s = "'" + strings.Replace(s, "'", "''", -1) + "'"
	case stringDoubleQuoted:
		fallthrough
	default:
		s = strconv.Quote(s)
	}
	enc.writeString(s)
}

func (enc *Encoder) write(data []byte) {
	if enc.err != nil {
		return
	}
	_, enc.err = enc.w.Write(data)
}

func (enc *Encoder) writeString(s string) {
	if enc.err != nil {
		return
	}
	_, enc.err = enc.w.WriteString(s)
}

func (enc *Encoder) writeInt(v int64) {
	if enc.err != nil {
		return
	}
	data := strconv.FormatInt(v, 10)
	_, enc.err = enc.w.WriteString(data)
}

func (enc *Encoder) writeFloat(v float64) {
	if enc.err != nil {
		return
	}
	data := strconv.FormatFloat(v, 'g', -1, 64)
	_, enc.err = enc.w.WriteString(data)
}

func (enc *Encoder) printf(format string, args ...interface{}) {
	if enc.err != nil {
		return
	}
	_, enc.err = fmt.Fprintf(enc.w, format, args...)
}

type stringFormat int

const (
	stringDoubleQuoted = stringFormat(iota)
	stringQuoted
	stringPlain
)

func bestStringFormat(s string) stringFormat {
	letters := true
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return stringDoubleQuoted
		} else if !unicode.IsLetter(r) {
			letters = false
		}
	}
	if !letters {
		return stringQuoted
	}
	if len(s) <= 5 {
		switch l := strings.ToLower(s); l {
		// http://yaml.org/type/null.html
		case "null", "~":
			return stringQuoted
		// http://yaml.org/type/bool.html
		case "true", "yes", "y", "on",
			"false", "no", "n", "off":
			return stringQuoted
		// http://yaml.org/type/merge.html
		case "<<":
			return stringQuoted
		}
	}
	return stringPlain
}

func Unmarshal(data []byte) (uast.Node, error) {
	var o interface{}
	if err := yaml.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	var fix func(interface{}) interface{}
	fix = func(o interface{}) interface{} {
		switch o := o.(type) {
		case map[interface{}]interface{}:
			m := make(map[string]interface{}, len(o))
			for k, v := range o {
				m[k.(string)] = fix(v)
			}
			return m
		case []interface{}:
			for i := range o {
				o[i] = fix(o[i])
			}
		}
		return o
	}
	o = fix(o)
	return uast.ToNode(o)
}