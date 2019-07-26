package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"bitbucket.org/creachadair/shell"
	"github.com/google/go-github/v27/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
)

// TODO(lwsanty): refactor

const (
	driversConfigURL = "https://raw.githubusercontent.com/bblfsh/documentation/master/languages.json"
	commitMsg        = "autogenerated changes"
	org              = "bblfsh"
)

var (
	errFailedToGetDrivers  = errors.NewKind("failed to get language drivers info")
	errFailedPrepareBranch = errors.NewKind("failed to prepare branch for driver %v: %v")
	errFailedPreparePR     = errors.NewKind("failed to prepare pull request for driver %v branch %v: %v")
	errCmdFailed           = errors.NewKind("command failed: %v, output: %v")
)

type Driver struct {
	Language string `json:"Language"`
	URL      string `json:"GithubURL"`
}

func GetDrivers() ([]Driver, error) {
	resp, err := http.Get(driversConfigURL)
	if err != nil {
		return nil, errFailedToGetDrivers.Wrap(err)
	}
	defer resp.Body.Close()

	var result []Driver
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errFailedToGetDrivers.Wrap(err)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, errFailedToGetDrivers.Wrap(err)
	}

	return result, nil
}

func PrepareBranch(d Driver, branch, script string) error {
	tmpDir, err := ioutil.TempDir("", d.Language)
	if err != nil {
		return err
	}
	log.Debugf("Created temp directory %v", tmpDir)
	defer func() { os.RemoveAll(tmpDir) }()

	origin := d.URL
	origin = strings.Replace(origin, "github.com", os.Getenv("APPLICATION")+":"+os.Getenv("TOKEN")+"@github.com", -1) + ".git"

	for _, cmnd := range []struct {
		logFormat string
		logArgs   []interface{}
		command   string
	}{
		{
			logFormat: "performing git clone repository %v -> %v",
			logArgs:   []interface{}{d.URL, tmpDir},
			command: fmt.Sprintf("git clone %[1]s %[2]s ; cd %[2]s ; "+
				"git remote rm origin ; git remote add origin %[3]s", shell.Quote(d.URL), shell.Quote(tmpDir), shell.Quote(origin)),
		},
		{
			logFormat: "creating branch %v",
			logArgs:   []interface{}{branch},
			command:   fmt.Sprintf("cd %s ; git checkout -b %s", shell.Quote(tmpDir), shell.Quote(branch)),
		},
		{
			logFormat: "executing the script",
			logArgs:   []interface{}{},
			command:   fmt.Sprintf("cd %s ; %v", shell.Quote(tmpDir), script),
		},
		{
			logFormat: "committing the changes",
			logArgs:   []interface{}{},
			command:   fmt.Sprintf("cd %s ; git add -A ; git commit --signoff -m \"%s\"", shell.Quote(tmpDir), commitMsg),
		},
		{
			logFormat: "pushing changes",
			logArgs:   []interface{}{},
			command:   fmt.Sprintf("cd %s ; git push origin %s", shell.Quote(tmpDir), shell.Quote(branch)),
		},
	} {
		log.Infof(cmnd.logFormat, cmnd.logArgs...)
		if err := ExecCmd(cmnd.command); err != nil {
			return errFailedPrepareBranch.New(d.Language, err)
		}
	}

	log.Infof("driver %v: branch %v has been successfully created", d.Language, branch)
	return nil
}

func PreparePR(d Driver, branch string) error {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("TOKEN")},
	)))

	pr, _, err := client.PullRequests.Create(ctx, org, d.Language+"-driver", &github.NewPullRequest{
		Title:               &branch,
		Head:                &branch,
		Base:                strPtr("master"),
		Body:                strPtr(commitMsg),
		MaintainerCanModify: newTrue(),
	})
	if err != nil {
		return errFailedPreparePR.New(d.Language, branch, err)
	}

	log.Infof("driver %v: pull request %v has been successfully created", d.Language, *pr.ID)
	return nil
}

// ExecCmd executes the specified Bash script. If execution fails, the error contains
// the combined output from stdout and stderr of the script.
// Do not use this for scripts that produce a large volume of output.
func ExecCmd(command string) error {
	cmd := exec.Command("bash", "-c", command)

	data, err := cmd.CombinedOutput()
	log.Debugf("command output: %v", string(data))
	if err != nil {
		return errCmdFailed.New(err, string(data))
	}

	return nil
}

func strPtr(s string) *string {
	return &s
}

func newTrue() *bool {
	b := true
	return &b
}
