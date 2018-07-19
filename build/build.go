package build

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/bblfsh/sdk.v2/assets/build"
	"gopkg.in/bblfsh/sdk.v2/internal/docker"
	"gopkg.in/bblfsh/sdk.v2/manifest"
	"gopkg.in/yaml.v2"
)

const releaseManifest = ".manifest.release.toml"

const (
	dockerFileName = docker.FileName
	manifestName   = "build.yml"
	ScriptName     = dockerFileName
)

func Verbose() bool {
	return isCI()
}

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

func (d *Driver) readManifest() (*manifest.Manifest, error) {
	return manifest.Load(d.path(manifest.Filename))
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
	Language string `yaml:"-"`
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
	if dm, err := d.readManifest(); err != nil {
		return nil, err
	} else {
		m.Language = dm.Language
	}
	if m.Native.Build.Gopath == "" && m.Native.Build.Image == "" {
		// if it's not a go build and build image is not specified - use native runtime image
		m.Native.Build.Image = m.Native.Image
	}
	return &m, nil
}
func (d *Driver) generateBuildScript() ([]byte, error) {
	m, err := d.readBuildManifest()
	if err != nil {
		return nil, err
	}

	text := string(build.MustAsset(dockerFileName + ".tpl"))
	tmpl := template.Must(template.New("").Parse(text))

	buf := bytes.NewBuffer(nil)
	err = tmpl.Execute(buf, m)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
func (d *Driver) Prepare() (bool, error) {
	if err := d.depEnsure(); err != nil {
		return false, err
	}

	data, err := d.generateBuildScript()
	if err != nil {
		return false, err
	}

	old, err := ioutil.ReadFile(d.path(dockerFileName))
	if err == nil && bytes.Equal(data, old) {
		return false, nil
	}

	out, err := create(d.path(dockerFileName))
	if err != nil {
		return false, err
	}
	defer out.Close()

	_, err = out.Write(data)
	if err != nil {
		return false, err
	}
	return true, out.Close()
}

func (d *Driver) ScriptChanged() (bool, error) {
	data, err := d.generateBuildScript()
	if err != nil {
		return false, err
	}
	old, err := ioutil.ReadFile(d.path(dockerFileName))
	if err != nil {
		return false, err
	}
	return !bytes.Equal(data, old), nil
}

func readYML(path string, dst interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, dst)
}

func (d *Driver) Build(imageName string) (string, error) {
	if _, err := d.Prepare(); err != nil {
		return "", err
	}
	if err := d.FillManifest(""); err != nil {
		return "", err
	}
	defer os.Remove(releaseManifest)

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

func (d *Driver) VersionTag() (string, error) {
	if vers := os.Getenv("DRIVER_VERSION"); vers != "" {
		return vers, nil
	} else if vers = ciTag(); vers != "" {
		return vers, nil
	}
	const devPrefix = "dev"
	tag := devPrefix
	rev, err := gitRev(d.root)
	if err != nil {
		return tag, err
	}
	tag += "-" + rev[:8]
	dirty, err := gitHasChanges(d.root)
	if err != nil {
		return tag, err
	}
	if dirty {
		tag += "-dirty"
	}
	return tag, nil
}

func (d *Driver) FillManifest(dest string) error {
	vers, err := d.VersionTag()
	if err != nil {
		return err
	}
	m, err := d.readManifest()
	if err != nil {
		return err
	}
	m.Version = vers

	now := time.Now().UTC()
	m.Build = &now

	bm, err := d.readBuildManifest()
	if err != nil {
		return err
	}
	m.Runtime.GoVersion = bm.Runtime.Version
	m.Runtime.NativeVersion = []string{bm.Native.Image}

	if dest == "" {
		dest = d.path(releaseManifest)
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	f, err := create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	return m.Encode(f)
}
