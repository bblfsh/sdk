package build

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/bblfsh/sdk/v3/assets/build"
	"github.com/bblfsh/sdk/v3/driver/manifest"
	"github.com/bblfsh/sdk/v3/internal/buildmanifest"
	"github.com/bblfsh/sdk/v3/internal/docker"
)

const releaseManifest = ".manifest.release.toml"

const (
	dockerFileName = docker.FileName
	manifestName   = buildmanifest.Filename
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

type buildManifest struct {
	Language string
	buildmanifest.Manifest
}

func (d *Driver) readBuildManifest() (*buildManifest, error) {
	data, err := ioutil.ReadFile(d.path(manifestName))
	if err != nil {
		return nil, err
	}
	var m buildManifest
	if err := m.Manifest.Decode(data); err != nil {
		return nil, err
	} else if m.Format != buildmanifest.CurrentFormat {
		return nil, fmt.Errorf("unknown format: %q", m.Format)
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
	for _, s := range m.Native.Deps {
		if !isApkOrApt(s) {
			return nil, fmt.Errorf("only apt/apk commands allowed in deps for final image")
		}
	}
	return &m, nil
}

func isApkOrApt(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	ok := false
	for _, p := range []string{
		"apt", "apt-get", "apk",
	} {
		if strings.HasPrefix(s, p+" ") {
			ok = true
		}
	}
	if !ok {
		return false
	}
	for _, sep := range []string{
		"&&", "&", "||", ";", "\n",
	} {
		sub := strings.Split(s, sep)
		if len(sub) == 1 {
			continue
		}
		for _, s := range sub {
			if !isApkOrApt(s) {
				return false
			}
		}
	}
	return true
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
	if err := d.modPrepare(); err != nil {
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
	var out io.Writer = buf
	if Verbose() {
		out = io.MultiWriter(buf, os.Stderr)
	}
	err = cli.BuildImage(docker.BuildImageOptions{
		Name:           imageName,
		ContextDir:     d.root,
		Dockerfile:     docker.FileName,
		SuppressOutput: true,
		OutputStream:   out,
		ErrorStream:    os.Stderr,
	})
	if err != nil {
		if !Verbose() {
			buf.WriteTo(os.Stderr)
		}
		return "", err
	}
	id := string(bytes.TrimSpace(buf.Bytes()))
	if strings.Contains(id, " ") {
		return "", fmt.Errorf("cannot parse container id: %q", id)
	}
	return id, nil
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
	if len(rev) >= 8 {
		rev = rev[:8]
	} else if rev == "" {
		rev = "unknown"
	}
	tag += "-" + rev
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
	m.Runtime.NativeVersion = bm.Native.Image

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
