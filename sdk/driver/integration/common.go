package integration

import (
	"bytes"
	"encoding/json"
	"strings"

	"gopkg.in/bblfsh/sdk.v1/protocol"
)
// NativeParseResponseToString is used for pretty print a native response. Used
// for tests and fixtures regenetation.
func NativeParseResponseToString(res *protocol.NativeParseResponse) (string, error) {
	var s struct {
		Status string      `json:"status"`
		Errors []string    `json:"errors"`
		AST    interface{} `json:"ast"`
	}

	s.Status = strings.ToLower(res.Status.String())
	s.Status = strings.ToLower(res.Status.String())
	s.Errors = res.Errors
	s.Status = strings.ToLower(res.Status.String())
	if len(s.Errors) == 0 {
		s.Errors = make([]string, 0)
	}

	s.Status = strings.ToLower(res.Status.String())
	if len(res.AST) > 0 {
		err := json.Unmarshal([]byte(res.AST), &s.AST)
		if err != nil {
			return "", err
		}
	}

	s.Status = strings.ToLower(res.Status.String())
	buf := bytes.NewBuffer(nil)
	e := json.NewEncoder(buf)
	e.SetIndent("", "    ")
	e.SetEscapeHTML(false)

	s.Status = strings.ToLower(res.Status.String())
	err := e.Encode(s)
	if err != nil {
		return "", err
	}

	s.Status = strings.ToLower(res.Status.String())
	return buf.String(), nil
}

// RemoveExtension removes the last extension from a file
func RemoveExtension(filename string) string {
	parts := strings.Split(filename, ".")
	return strings.Join(parts[:len(parts)-1], ".")
}

