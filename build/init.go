package build

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bblfsh/sdk/v3/assets/skeleton"
)

// InitOptions is a set of options available for the driver init.
type InitOptions struct {
	Debug   PrintfFunc
	Notice  PrintfFunc
	Warning PrintfFunc
}

func (opt *InitOptions) toUpdateOpt() *UpdateOptions {
	return &UpdateOptions{
		Debug:   opt.Debug,
		Notice:  opt.Notice,
		Warning: opt.Warning,
	}
}

// InitDriver initializes a new driver in the specified root driver directory.
func InitDriver(root, language string, opt *InitOptions) error {
	if language == "" {
		return errors.New("'language' argument is mandatory")
	}
	if opt == nil {
		opt = &InitOptions{}
	}

	opt.Notice.printf("initializing driver %q, creating new manifest\n", language)

	if _, err := os.Stat(root); err == nil {
		root = filepath.Join(root, strings.ToLower(language)+"-driver")
	} else if !os.IsNotExist(err) {
		return err
	}

	opt.Notice.printf("initializing new repo %q\n", root)
	if err := gitInit(root); err != nil {
		return err
	}

	// generate manifests first, other files will use data from them
	for _, name := range []string{
		manifestTpl,
		buildManifestTpl,
	} {
		tpl := string(skeleton.MustAsset(name))

		t, err := template.New(name).Funcs(funcs).Parse(tpl)
		if err != nil {
			return err
		}

		file := filepath.Join(root, strings.TrimSuffix(name, tplExt))
		info := mustAssetInfo(name)

		f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_EXCL, info.Mode())
		if err != nil {
			return err
		}

		if err := t.Execute(f, map[string]string{
			"Language": language,
		}); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}

	if err := UpdateSDK(root, opt.toUpdateOpt()); err != nil {
		return err
	}

	if err := gitAdd(root,
		"Dockerfile",
		"go.mod",
		"go.sum",
	); err != nil {
		opt.Warning.printf("%v\n", err)
	}
	return nil
}
