package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bblfsh/sdk/v3/driver/manifest/discovery"

	"bitbucket.org/creachadair/shell"
	"github.com/google/go-github/v27/github"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
)

const (
	org       = "bblfsh"
	tmpFolder = "/var/lib/tmp"

	gitUser = "bblfsh-release-bot"
	gitMail = "<release-bot@bblf.sh>"

	errSpecialText = "nothing to commit"
)

var (
	errFailedToPrepareBranch = errors.NewKind("failed to prepare branch for driver %v: %v")
	errFailedToPreparePR     = errors.NewKind("failed to prepare pull request for driver %v branch %v: %v")
	errCmdFailed             = errors.NewKind("command failed: %v, output: %v")
	// ErrNothingToCommit is a specific error that should not stop the global process of drivers update
	ErrNothingToCommit = errors.NewKind(errSpecialText)
)

type pipeLine struct {
	nodes      []pipeLineNode
	driver     discovery.Driver
	dockerfile bool
	dryRun     bool
	tmpDir     string
}

type pipeLineNode struct {
	logFormat string
	logArgs   []interface{}
	command   string
}

// UpdateOptions represents git metadata for changes and ways of execution of update script
type UpdateOptions struct {
	Branch              string
	SDKVersion          string
	Script              string
	CommitMsg           string
	Dockerfile          bool
	ExplicitCredentials bool
	DryRun              bool
}

func newPipeLine(d discovery.Driver, githubToken string, o *UpdateOptions) *pipeLine {
	url := d.RepositoryURL()
	processOptions(o)
	origin := getOrigin(url, githubToken, o)
	tmpDir := filepath.Join(tmpFolder, d.Language)

	var nodes []pipeLineNode
	nodes = append(nodes,
		pipeLineNode{
			logFormat: "creating dir %v",
			logArgs:   []interface{}{tmpDir},
			command:   fmt.Sprintf("mkdir -p %v", shell.Quote(tmpDir)),
		},
		pipeLineNode{
			logFormat: "performing git clone repository %v -> %v",
			logArgs:   []interface{}{url, tmpDir},
			command: fmt.Sprintf("git clone %[1]s %[2]s ; cd %[2]s ; "+
				"git remote rm origin ; git remote add origin %[3]s", shell.Quote(url), shell.Quote(tmpDir), origin),
		}, pipeLineNode{
			logFormat: "creating branch %v",
			logArgs:   []interface{}{o.Branch},
			command:   fmt.Sprintf("cd %s ; git checkout -b %s", shell.Quote(tmpDir), shell.Quote(o.Branch)),
		})
	if strings.TrimSpace(o.Script) != "" {
		script := o.Script
		if o.Dockerfile {
			script = strings.Replace(script, "\n", ";", -1)
		}
		nodes = append(nodes, pipeLineNode{
			logFormat: "executing the script",
			logArgs:   []interface{}{},
			command:   fmt.Sprintf("cd %s ; %v", shell.Quote(tmpDir), script),
		})
	}
	if strings.TrimSpace(o.SDKVersion) != "" {
		nodes = append(nodes, pipeLineNode{
			logFormat: "updating sdk to %v",
			logArgs:   []interface{}{o.SDKVersion},
			command:   fmt.Sprintf("cd %s ; go mod download ; go mod edit -require github.com/bblfsh/sdk/v3@v%v ; go run ./update.go", shell.Quote(tmpDir), o.SDKVersion),
		})
	}
	nodes = append(nodes, pipeLineNode{
		logFormat: "set git user info",
		logArgs:   []interface{}{},
		command:   fmt.Sprintf("cd %s ; git config --global user.name %v ; git config --global user.email %v", shell.Quote(tmpDir), gitUser, shell.Quote(gitMail)),
	}, pipeLineNode{
		logFormat: "committing the changes",
		logArgs:   []interface{}{},
		command:   fmt.Sprintf("cd %s ; git add -A ; git commit --signoff -m \"%s\"", shell.Quote(tmpDir), o.CommitMsg),
	}, pipeLineNode{
		logFormat: "pushing changes",
		logArgs:   []interface{}{},
		command:   fmt.Sprintf("cd %s ; git push origin %s", shell.Quote(tmpDir), shell.Quote(o.Branch)),
	})

	return &pipeLine{nodes: nodes, driver: d, dockerfile: o.Dockerfile, tmpDir: tmpDir, dryRun: o.DryRun}
}

func (p *pipeLine) createDockerfile() (string, error) {
	const header = `FROM golang:1.12

ARG GITHUB_TOKEN

`
	content := header
	for _, c := range p.nodes {
		comment := fmt.Sprintf(c.logFormat, c.logArgs...)
		content += fmt.Sprintf("# %v\nRUN %v\n\n", comment, c.command)
	}

	_, err := os.Stat(p.tmpDir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(p.tmpDir, os.ModePerm); err != nil {
			return "", err
		}
	}

	path := filepath.Join(p.tmpDir, "Dockerfile")
	log.Infof("trying to create: %v", path)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	log.Infof("preparing dockerfile:\n%v", content)
	if _, err := f.WriteString(content); err != nil {
		return "", err
	}

	return path, nil
}

func (p *pipeLine) exec(githubToken string) error {
	if p.dockerfile {
		dockerPath, err := p.createDockerfile()
		if err != nil {
			return err
		}
		if p.dryRun {
			return nil
		}
		command := fmt.Sprintf("docker build --build-arg GITHUB_TOKEN=%v -t %v-driver-update %v",
			githubToken, p.driver.Language, filepath.Dir(dockerPath))
		if err := ExecCmd(command); err != nil {
			err = errFailedToPrepareBranch.New(p.driver.Language, err)
			if strings.Contains(err.Error(), errSpecialText) {
				err = ErrNothingToCommit.Wrap(err)
			}
			return err
		}
		return nil
	}

	for _, c := range p.nodes {
		log.Infof(c.logFormat, c.logArgs...)
		if p.dryRun {
			log.Infof(c.command)
			continue
		}
		if err := ExecCmd(c.command); err != nil {
			return errFailedToPrepareBranch.New(p.driver.Language, err)
		}
	}
	return nil
}

func (p *pipeLine) close() {
	if err := os.RemoveAll(p.tmpDir); err != nil {
		log.Warningf("could not remove directory %v: %v", p.tmpDir, err)
	}
}

// PrepareBranch does the next steps:
// 1) clones driver's master branch
// 2) creates new branch
// 3) executes custom script if it's not empty
// 4) updates SDK version if it's not empty
// 5) commits and pushes changes to the previously created branch
func PrepareBranch(d discovery.Driver, githubToken string, o *UpdateOptions) error {
	p := newPipeLine(d, githubToken, o)
	defer p.close()
	if err := p.exec(githubToken); err != nil {
		return err
	}

	log.Infof("driver %v: branch %v has been successfully created", d.Language, o.Branch)
	return nil
}

// PreparePR creates pull request for a given driver's branch
func PreparePR(d discovery.Driver, githubToken, branch, commitMsg string, dryRun bool) error {
	ctx := context.Background()
	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)))

	log.Infof("Preparing pr %v -> master", branch)
	newPR := &github.NewPullRequest{
		Title:               &branch,
		Head:                &branch,
		Base:                strPtr("master"),
		Body:                strPtr(commitMsg),
		MaintainerCanModify: newTrue(),
	}
	if dryRun {
		log.Infof("pr to be created:\ntitle: %v\nhead: %v\nbase: %v\nbody: %v\nmaintainers can modify: %v",
			newPR.GetTitle(), newPR.GetHead(), newPR.GetBase(), newPR.GetBody(), newPR.GetMaintainerCanModify())
		return nil
	}

	pr, _, err := client.PullRequests.Create(ctx, org, d.Language+"-driver", newPR)
	if err != nil {
		return errFailedToPreparePR.New(d.Language, branch, err)
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

func getOrigin(url string, githubToken string, o *UpdateOptions) string {
	token := githubToken
	if o.Dockerfile && !o.ExplicitCredentials {
		token = "${GITHUB_TOKEN}"
	}
	return strings.Replace(url, "github.com", gitUser+":"+token+"@github.com", -1)
}

func processOptions(o *UpdateOptions) {
	o.Branch = strings.TrimSpace(o.Branch)
	o.CommitMsg = strings.TrimSpace(o.CommitMsg)
	o.SDKVersion = strings.TrimSpace(o.SDKVersion)
}

func strPtr(s string) *string {
	return &s
}

func newTrue() *bool {
	b := true
	return &b
}
