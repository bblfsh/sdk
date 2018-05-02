package fixtures

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/bblfsh/sdk.v2/protocol"
	"gopkg.in/bblfsh/sdk.v2/sdk/driver"
	"gopkg.in/bblfsh/sdk.v2/uast"
	"gopkg.in/yaml.v2"
)

const Dir = "fixtures"

type Suite struct {
	Lang     string
	Ext      string // with dot
	Path     string
	WriteYML bool

	// Update* flags below should never be committed in "true" state.
	// They serve only as an automation for updating fixture files.

	UpdateNative bool // update native ASTs in fixtures to ones produced by driver
	UpdateUAST   bool // update UASTs in fixtures to ones produced by driver

	NewDriver  func() driver.BaseDriver
	Transforms driver.Transforms

	BenchName string // fixture name to benchmark (with no extension)
}

func (s *Suite) readFixturesFile(t testing.TB, name string) string {
	data, err := ioutil.ReadFile(filepath.Join(s.Path, name))
	if os.IsNotExist(err) {
		return ""
	}
	require.NoError(t, err)
	return string(data)
}

func (s *Suite) writeFixturesFile(t testing.TB, name string, data string) {
	err := ioutil.WriteFile(filepath.Join(s.Path, name), []byte(data), 0644)
	require.NoError(t, err)
}

func (s *Suite) deleteFixturesFile(name string) {
	os.Remove(filepath.Join(s.Path, name))
}

func (s *Suite) RunTests(t *testing.T) {
	t.Run("native", s.testFixturesNative)
	t.Run("uast", s.testFixturesUAST)
}

func (s *Suite) RunBenchmarks(t *testing.B) {
	t.Run("transform", s.benchmarkTransform)
}

const (
	gotSuffix = "2"
	nativeExt = ".native"
	uastExt   = ".uast"
)

func (s *Suite) testFixturesNative(t *testing.T) {
	list, err := ioutil.ReadDir(s.Path)
	require.NoError(t, err)

	dr := s.NewDriver()

	err = dr.Start()
	require.NoError(t, err)
	defer dr.Close()

	suffix := s.Ext
	for _, ent := range list {
		if !strings.HasSuffix(ent.Name(), suffix) {
			continue
		}
		name := strings.TrimSuffix(ent.Name(), suffix)
		t.Run(name, func(t *testing.T) {
			code := s.readFixturesFile(t, name+suffix)

			resp, err := dr.Parse(&driver.InternalParseRequest{
				Content:  string(code),
				Encoding: driver.Encoding(protocol.UTF8),
			})
			require.NoError(t, err)

			if s.WriteYML {
				ya, err := yaml.Marshal(resp.AST)
				require.NoError(t, err)
				s.writeFixturesFile(t, name+suffix+nativeExt+".yml", string(ya))
			}

			js, err := json.Marshal(resp.AST)
			require.NoError(t, err)

			exp := s.readFixturesFile(t, name+suffix+nativeExt)
			got := (&protocol.NativeParseResponse{
				Response: protocol.Response{
					Status: protocol.Status(resp.Status),
					Errors: resp.Errors,
				},
				AST:      string(js),
				Language: s.Lang,
			}).String()
			if exp == "" {
				s.writeFixturesFile(t, name+suffix+nativeExt, got)
				t.Skip("no test file found - generating")
			}
			if !assert.ObjectsAreEqual(exp, got) {
				ext := nativeExt + gotSuffix
				if s.UpdateNative {
					ext = nativeExt
				}
				s.writeFixturesFile(t, name+suffix+ext, got)
			} else {
				s.deleteFixturesFile(name + suffix + nativeExt + gotSuffix)
			}
			require.Equal(t, exp, got)
		})
	}
}

func (s *Suite) testFixturesUAST(t *testing.T) {
	list, err := ioutil.ReadDir(s.Path)
	require.NoError(t, err)

	dr := s.NewDriver()

	err = dr.Start()
	require.NoError(t, err)
	defer dr.Close()

	suffix := s.Ext
	for _, ent := range list {
		if !strings.HasSuffix(ent.Name(), suffix) {
			continue
		}
		name := strings.TrimSuffix(ent.Name(), suffix)
		t.Run(name, func(t *testing.T) {
			code := s.readFixturesFile(t, name+suffix)

			req := &driver.InternalParseRequest{
				Content:  string(code),
				Encoding: driver.Encoding(protocol.UTF8),
			}

			resp, err := dr.Parse(req)
			require.NoError(t, err)

			ast, err := uast.ToNode(resp.AST)
			require.NoError(t, err)

			tr := s.Transforms
			ua, err := tr.Do(driver.ModeAST, code, ast)
			require.NoError(t, err)

			if s.WriteYML {
				ya, err := yaml.Marshal(ua)
				require.NoError(t, err)
				s.writeFixturesFile(t, name+suffix+uastExt+".yml", string(ya))
			}

			un, err := protocol.ToNode(ua)
			require.NoError(t, err)

			exp := s.readFixturesFile(t, name+suffix+uastExt)
			got := (&protocol.ParseResponse{
				Response: protocol.Response{
					Status: protocol.Status(resp.Status),
					Errors: resp.Errors,
				},
				UAST:     un,
				Language: s.Lang,
			}).String()
			if exp == "" {
				s.writeFixturesFile(t, name+suffix+uastExt, got)
				t.Skip("no test file found - generating")
			}
			if !assert.ObjectsAreEqual(exp, got) {
				ext := uastExt + gotSuffix
				if s.UpdateUAST {
					ext = uastExt
				}
				s.writeFixturesFile(t, name+suffix+ext, got)
			} else {
				s.deleteFixturesFile(name + suffix + uastExt + gotSuffix)
			}
			require.Equal(t, exp, got)
		})
	}
}

func (s *Suite) benchmarkTransform(b *testing.B) {
	if s.BenchName == "" {
		b.SkipNow()
	}
	code := s.readFixturesFile(b, s.BenchName+s.Ext)
	req := &driver.InternalParseRequest{
		Content:  string(code),
		Encoding: driver.Encoding(protocol.UTF8),
	}

	tr := s.Transforms

	dr := s.NewDriver()

	err := dr.Start()
	require.NoError(b, err)
	defer dr.Close()

	resp, err := dr.Parse(req)
	if err != nil {
		b.Fatal(err)
	}
	dr.Close()

	rast, err := uast.ToNode(resp.AST)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ast := rast.Clone()

		ua, err := tr.Do(driver.ModeAST, code, ast)
		if err != nil {
			b.Fatal(err)
		}

		un, err := protocol.ToNode(ua)
		if err != nil {
			b.Fatal(err)
		}
		_ = un
	}
}
