package build

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/bblfsh/sdk.v2/assets/build"
	"gopkg.in/yaml.v2"
)

const (
	dockerFileName = "Dockerfile"
	manifestName   = "build.yml"
)

type artifact struct {
	Path string `yaml:"path"`
	Dest string `yaml:"dest"`
}

type manifest struct {
	SDK      string `yaml:"sdk"`
	Language string `yaml:"language"`
	Native   struct {
		Image  string     `yaml:"image"`
		Static []artifact `yaml:"static"`
		Build  struct {
			Gopath    string     `yaml:"gopath"`
			Image     string     `yaml:"image"`
			Deps      []string   `yaml:"deps"`
			Add       []artifact `yaml:"add"`
			Run       []string   `yaml:"run"`
			Artifacts []artifact `yaml:"artifacts"`
		} `yaml:"build"`
		Test struct {
			Deps []string `yaml:"deps"`
			Run  []string `yaml:"run"`
		} `yaml:"test"`
	} `yaml:"native"`
	Runtime struct {
		Version string `yaml:"version"`
	} `yaml:"go-runtime"`
}

func Prepare(path string) error {
	if err := depEnsure(path); err != nil {
		return err
	}

	var m manifest
	if err := readYML(filepath.Join(path, manifestName), &m); err != nil {
		return err
	} else if m.SDK != "2" {
		return fmt.Errorf("unknown SDK version: %q", m.SDK)
	}
	if m.Native.Build.Gopath == "" && m.Native.Build.Image == "" {
		// if it's not a go build and build image is not specified - use native runtime image
		m.Native.Build.Image = m.Native.Image
	}

	text := string(build.MustAsset(dockerFileName + ".tpl"))
	tmpl := template.Must(template.New("").Parse(text))

	out, err := create(filepath.Join(dockerFileName))
	if err != nil {
		return err
	}
	defer out.Close()

	return tmpl.Execute(out, m)
}

func readYML(path string, dst interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, dst)
}

func Build(path string, name string) (string, error) {
	if err := Prepare(path); err != nil {
		return "", err
	}
	args := []string{"build", "-q"}
	if name != "" {
		args = append(args, "-t", name)
	}
	args = append(args, ".")

	buf := bytes.NewBuffer(nil)
	err := execIn(path, buf, "docker", args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
