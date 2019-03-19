package build

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"gopkg.in/bblfsh/sdk.v2/assets/skeleton"
	"gopkg.in/bblfsh/sdk.v2/driver/manifest"
)

const (
	tplExt      = ".tpl"
	manifestTpl = manifest.Filename + tplExt
)

func genEnvBool(key string) bool {
	str := os.Getenv(key)
	if str == "" {
		return false
	}
	v, err := strconv.ParseBool(str)
	if err != nil {
		panic(err)
	}
	return v
}

var overwriteManagedFiles = genEnvBool("BABELFISH_OVERWRITE_MANAGED")

// managedFiles are files that always are overwritten
var managedFiles = map[string]bool{
	".travis.yml":             true,
	"Makefile":                true,
	"update.go":               true,
	"build.go":                true,
	"test.go":                 true,
	"README.md" + tplExt:      true,
	"LICENSE":                 true,
	"driver/main.go" + tplExt: true,
	"driver/sdk_test.go":      true,
	"driver/normalizer/transforms.go" + tplExt: true,
}

type updater struct {
	root    string
	opt     UpdateOptions
	context map[string]interface{}
	changes int
}

// PrintfFunc is a logging function type similar to log.Printf.
type PrintfFunc func(format string, args ...interface{}) (int, error)

func (f PrintfFunc) printf(msg string, args ...interface{}) {
	if f == nil {
		return
	}
	_, _ = f(msg, args...)
}

func mustAssetInfo(name string) os.FileInfo {
	fi, err := skeleton.AssetInfo(name)
	if err != nil {
		panic(fmt.Errorf("missing asset info for %q: %v", name, err))
	}
	return fi
}

// UpdateOptions is a set of options available for
type UpdateOptions struct {
	DryRun bool

	Debug   PrintfFunc
	Notice  PrintfFunc
	Warning PrintfFunc
}

// ErrChangesRequired is returned by UpdateSDK in DryRun mode when changes are required.
var ErrChangesRequired = errors.New("changes are required")

// GenerateManifest writes the manifest file to a root driver directory.
func GenerateManifest(root, language string) error {
	tpl := string(skeleton.MustAsset(manifestTpl))

	t, err := template.New(manifestTpl).Funcs(funcs).Parse(tpl)
	if err != nil {
		return err
	}

	name := fixGitFolder(manifestTpl)
	file := filepath.Join(root, manifest.Filename)

	info := mustAssetInfo(name)

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_EXCL, info.Mode())
	if err != nil {
		return err
	}
	defer f.Close()

	if err := t.Execute(f, map[string]string{
		"Language": language,
	}); err != nil {
		return err
	}
	return f.Close()
}

// UpdateSDK updates SDK-managed files for the driver located at root.
//
// If DryRun option is set, the function would not update any files, and instead will
// return ErrChangesRequired if there are any changes required.
func UpdateSDK(root string, opt *UpdateOptions) error {
	if opt == nil {
		opt = &UpdateOptions{}
	}

	m, err := manifest.Load(filepath.Join(root, manifest.Filename))
	if err != nil {
		return err
	}
	c := &updater{
		root: root, opt: *opt,
		context: map[string]interface{}{
			"Manifest": m,
		},
	}

	for _, file := range skeleton.AssetNames() {
		if err := c.processAsset(file); err != nil {
			return err
		}
	}
	d, err := NewDriver(root)
	if err != nil {
		return err
	}
	var changed bool
	if c.opt.DryRun {
		changed, err = d.ScriptChanged()
	} else {
		changed, err = d.Prepare()
	}
	if err != nil {
		return err
	} else if changed {
		c.notifyChangedFile(ScriptName)
	}

	if c.opt.DryRun && c.changes > 0 {
		return ErrChangesRequired
	}

	return nil
}

func (c *updater) processAsset(name string) error {
	if name == manifestTpl {
		// manifest is always managed by the driver developer
		// the template is only for the driver init
		return nil
	}
	overwrite := managedFiles[name] && !overwriteManagedFiles

	if strings.HasSuffix(name, tplExt) {
		return c.processTemplateAsset(name, c.context, overwrite)
	}

	return c.processFileAsset(name, overwrite)
}

func (c *updater) processFileAsset(name string, overwrite bool) error {
	content := skeleton.MustAsset(name)
	info := mustAssetInfo(name)

	name = fixGitFolder(name)
	return c.writeIfChanged(filepath.Join(c.root, name), content, info.Mode(), overwrite)
}

var funcs = map[string]interface{}{
	"escape_shield":  escapeShield,
	"toExportedName": toExportedName,
}

func (c *updater) processTemplateAsset(name string, v interface{}, overwrite bool) error {
	tpl := string(skeleton.MustAsset(name))

	t, err := template.New(name).Funcs(funcs).Parse(tpl)
	if err != nil {
		return err
	}

	name = fixGitFolder(name)
	file := filepath.Join(c.root, name[:len(name)-len(tplExt)])

	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, v); err != nil {
		return err
	}

	info := mustAssetInfo(name)
	return c.writeIfChanged(file, buf.Bytes(), info.Mode(), overwrite)
}

func (c *updater) writeIfChanged(file string, content []byte, m os.FileMode, overwrite bool) error {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		c.notifyMissingFile(file)
		return c.writeFile(file, content, m)
	} else if err != nil {
		return err
	} else if !overwrite {
		return nil
	}

	original, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	if bytes.Compare(original, content) == 0 {
		return nil
	}

	c.notifyChangedFile(file)
	return c.writeFile(file, content, m)
}

func (c *updater) writeFile(file string, content []byte, m os.FileMode) error {
	if c.opt.DryRun {
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return err
	}

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, m)
	if err != nil {
		return err
	}

	defer f.Close()

	if c.opt.Debug != nil {
		c.opt.Debug.printf("file %q has been written\n", file)
	}

	_, err = f.Write(content)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(c.root, file)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(rel, ".git"+string(filepath.Separator)) {
		git := exec.Command("git", "add", rel)
		git.Dir = c.root
		if out, err := git.CombinedOutput(); err != nil {
			c.opt.Warning.printf("cannot add a file to git: %v\n%s", err, string(out))
		}
	}
	return nil
}

func (c *updater) notifyMissingFile(file string) {
	if isDotGit(file) {
		return
	}

	if !c.opt.DryRun {
		c.opt.Notice.printf("creating file %q\n", file)
		return
	}

	c.changes++
	c.opt.Warning.printf("missing file %q\n", file)
}

func (c *updater) notifyChangedFile(file string) {
	if !c.opt.DryRun {
		c.opt.Warning.printf("managed file %q has changed, overriding changes\n", file)
		return

	}

	c.changes++
	c.opt.Warning.printf("managed file changed %q\n", file)
}

func escapeShield(text interface{}) string {
	return strings.Replace(fmt.Sprintf("%s", text), "-", "--", -1)
}

func toExportedName(s string) string {
	r, n := utf8.DecodeRuneInString(s)
	if n == 0 {
		return s
	}
	return string(unicode.ToUpper(r)) + s[n:]
}

func fixGitFolder(path string) string {
	return strings.Replace(path, "git/", ".git/", 1)
}

func isDotGit(path string) bool {
	return strings.Contains(path, ".git/")
}
