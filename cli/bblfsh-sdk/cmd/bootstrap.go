package cmd

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bblfsh/sdk/assets/skeleton"
	"github.com/bblfsh/sdk/manifest"
)

const (
	tplExtension = ".tpl"
	manifestTpl  = "manifest.toml.tpl"
)

// managedFiles are files that always are overwritten
var managedFiles = map[string]bool{
	".travis.yml":        true,
	"README.md.tpl":      true,
	"LICENSE":            true,
	"driver/main.go.tpl": true,
}

type BootstrapCommand struct {
	Verbose      bool   `short:"v" description:"show verbose debug information"`
	InitLanguage string `long:"init-lang" description:"create an initial manifest.toml, based on the given lang"`
	Args         struct {
		Root string `positional-arg-name:"project-root" default:"."`
	} `positional-args:"yes"`

	context map[string]interface{}
}

func (c *BootstrapCommand) Execute(args []string) error {
	if err := c.processManifest(); err != nil {
		return err
	}

	m, err := c.readManifest()
	if err != nil {
		return err
	}

	c.context = map[string]interface{}{
		"Manifest": m,
	}

	for _, file := range skeleton.AssetNames() {
		if file == manifestTpl {
			continue
		}

		if err := c.processAsset(file); err != nil {
			return err
		}
	}

	return nil
}

func (c *BootstrapCommand) processManifest() error {
	if c.InitLanguage == "" {
		return nil
	}

	notice.Printf("initializing driver %q, creating new manifest\n", c.InitLanguage)
	if _, err := c.readManifest(); err == nil {
		warning.Printf("driver already initialized. %q detected\n", manifest.Filename)
	}

	return c.processTemplateAsset(manifestTpl, map[string]interface{}{
		"Language": c.InitLanguage,
	}, false)

}

func (c *BootstrapCommand) processAsset(name string) error {
	overwrite := managedFiles[name]

	if strings.HasSuffix(name, tplExtension) {
		return c.processTemplateAsset(name, c.context, overwrite)
	}

	return c.processFileAsset(name, overwrite)
}
func (c *BootstrapCommand) processFileAsset(name string, overwrite bool) error {
	content := skeleton.MustAsset(name)
	return c.writeTemplate(filepath.Join(c.Args.Root, name), content, overwrite)
}

func (c *BootstrapCommand) processTemplateAsset(name string, v map[string]interface{}, overwrite bool) error {
	tpl := string(skeleton.MustAsset(name))
	t, err := template.New(name).Parse(tpl)
	if err != nil {
		return err
	}

	file := filepath.Join(c.Args.Root, name[:len(name)-len(tplExtension)])

	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, v); err != nil {
		return err
	}

	return c.writeTemplate(file, buf.Bytes(), overwrite)
}

func (c *BootstrapCommand) writeTemplate(file string, content []byte, overwrite bool) error {
	f, err := os.Open(file)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if f == nil {
		notice.Printf("creating file %q\n", file)
		return c.doWriteTemplate(file, content)
	}

	if !overwrite {
		return nil
	}

	original, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if bytes.Compare(original, content) == 0 {
		return nil
	}

	warning.Printf("managed file %q has changed, discarding changes\n", file)
	return c.doWriteTemplate(file, content)
}

func (c *BootstrapCommand) doWriteTemplate(file string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()

	if c.Verbose {
		debug.Printf("file %q has be written\n", file)
	}

	_, err = f.Write(content)
	return err
}

func (c *BootstrapCommand) readManifest() (*manifest.Manifest, error) {
	return manifest.Load(filepath.Join(c.Args.Root, manifest.Filename))
}
