package build

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/bblfsh/sdk.v2/assets/build"
	"gopkg.in/bblfsh/sdk.v2/internal/docker"
	"gopkg.in/yaml.v2"
)

const (
	dockerFileName = docker.FileName
	manifestName   = "build.yml"
)

func NewDriver(path string) (*Driver, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &Driver{root: path}, nil
}

type Driver struct {
	root string
}

func (d *Driver) path(names ...string) string {
	return filepath.Join(
		append([]string{d.root}, names...)...,
	)
}

type artifact struct {
	Path string `yaml:"path"`
	Dest string `yaml:"dest"`
}

type buildManifest struct {
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

func (d *Driver) readBuildManifest() (*buildManifest, error) {
	var m buildManifest
	if err := readYML(d.path(manifestName), &m); err != nil {
		return nil, err
	} else if m.SDK != "2" {
		return nil, fmt.Errorf("unknown SDK version: %q", m.SDK)
	}
	if m.Native.Build.Gopath == "" && m.Native.Build.Image == "" {
		// if it's not a go build and build image is not specified - use native runtime image
		m.Native.Build.Image = m.Native.Image
	}
	return &m, nil
}
func (d *Driver) Prepare() error {
	if err := d.depEnsure(); err != nil {
		return err
	}

	m, err := d.readBuildManifest()
	if err != nil {
		return err
	}

	text := string(build.MustAsset(dockerFileName + ".tpl"))
	tmpl := template.Must(template.New("").Parse(text))

	out, err := create(d.path(dockerFileName))
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

func (d *Driver) Build(imageName string) (string, error) {
	if err := d.Prepare(); err != nil {
		return "", err
	}

	cli, err := docker.Dial()
	if err != nil {
		return "", err
	}

	args := []string{"build", "-q"}
	if imageName != "" {
		args = append(args, "-t", imageName)
	}
	args = append(args, ".")
	printCommand("docker", args...)

	buf := bytes.NewBuffer(nil)
	err = cli.BuildImage(docker.BuildImageOptions{
		Name:           imageName,
		ContextDir:     d.root,
		Dockerfile:     docker.FileName,
		SuppressOutput: true,
		OutputStream:   buf,
		ErrorStream:    os.Stderr,
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}
