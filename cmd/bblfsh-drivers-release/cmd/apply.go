package cmd

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/bblfsh/sdk/v3/cmd"
	"github.com/bblfsh/sdk/v3/driver/manifest/discovery"

	"github.com/google/go-github/github"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
	"gopkg.in/yaml.v2"
)

const ApplyCommandDescription = "takes YAML file, generated with collect command and edited with new releases info and performs new releases if it's required"

var errFailedToCreateRelease = errors.NewKind("%v: failed to create release: %v")

type ApplyCommand struct {
	cmd.Command

	DryRun        bool   `long:"dry-run" description:"performs extra debug info instead of the real action"`
	File          string `long:"file" short:"f" env:"FILE" default:"drivers-releases.yml" description:"path to file with configuration"`
	ReleaseBranch string `long:"release-branch" env:"RELEASE_BRANCH" default:"master" description:"branch to release"`
}

func (c *ApplyCommand) Execute(args []string) error {
	ctx := context.Background()

	data, err := ioutil.ReadFile(c.File)
	if err != nil {
		return err
	}

	desiredReleases := make(DriversReleases)
	if err := yaml.Unmarshal(data, desiredReleases); err != nil {
		return err
	}
	actualReleases, err := getDriversReleases(ctx)
	if err != nil {
		return err
	}

	var lastErr error
	for language, release := range desiredReleases {
		if release.Tag == actualReleases[language].Tag {
			log.Warningf("%v-language: already tagged with %v release, please consider another iteration", language, release.Tag)
			continue
		}
		if err := c.releaseDriver(ctx, language+"-driver", release); err != nil {
			lastErr = err
			log.Warningf(err.Error())
		}
	}
	if lastErr != nil {
		log.Warningf("release failed for one or several drivers")
		os.Exit(1)
	}

	return nil
}

// releaseDriver uses github API to create the release from previously parsed config file
func (c *ApplyCommand) releaseDriver(ctx context.Context, repo string, release Release) error {
	client := discovery.GithubClient()

	releaseConfig := &github.RepositoryRelease{
		TagName:         github.String(release.Tag),
		TargetCommitish: github.String(c.ReleaseBranch),
		Name:            github.String(release.Tag),
		Body:            github.String(release.Description),
		Draft:           github.Bool(false),
		Prerelease:      github.Bool(false),
	}

	if c.DryRun {
		data, err := yaml.Marshal(releaseConfig)
		if err != nil {
			return err
		}
		log.Infof("%s-driver: release config:\n%v\n", repo, string(data))
		return nil
	}

	releaseResp, _, err := client.Repositories.CreateRelease(ctx, owner, repo, releaseConfig)
	if err != nil {
		return errFailedToCreateRelease.New(repo, err)
	}
	log.Infof("%v: release %v has been successfully created", repo, *releaseResp.Name)
	return nil
}
