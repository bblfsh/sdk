package cmd

import (
	"bufio"
	"context"
	"io/ioutil"
	"strings"

	"github.com/bblfsh/sdk/v3/cmd"
	"github.com/bblfsh/sdk/v3/driver/manifest/discovery"

	"github.com/google/go-github/github"
	"gopkg.in/src-d/go-errors.v1"
	"gopkg.in/src-d/go-log.v1"
	"gopkg.in/yaml.v2"
)

const (
	owner = discovery.GithubOrg

	CollectCommandDescription = "pulls info about all latest releases of drivers and all commits since those releases and dumps this info to a YAML file"
)

var (
	errFailedToGetCommitsDescription = errors.NewKind("%v: failed to get commits description: %v")
	errFailedToGetDriversReleases    = errors.NewKind("failed to get drivers releases")
)

// DriversReleases represents map with
// - key: language
// - value: Release object
type DriversReleases map[string]Release

// Release represents an object with the latest tag + concatenated descriptions of commits, performed after release
type Release struct {
	// Tag corresponds to release tag in format v0.0.1
	Tag string `yaml:"tag"`
	// Description represents concatenated descriptions of commits
	Description string `yaml:"description"`
}

type CollectCommand struct {
	cmd.Command

	DryRun bool   `long:"dry-run" description:"performs extra debug info instead of the real action"`
	File   string `long:"file" short:"f" env:"FILE" default:"drivers-releases.yml" description:"path to file with configuration"`
}

func (c *CollectCommand) Execute(args []string) error {
	drivers, err := getDriversReleases(context.Background())
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(drivers)
	if err != nil {
		return err
	}

	if c.DryRun {
		log.Infof("\n%v\n", string(data))
		return nil
	}
	if err := ioutil.WriteFile(c.File, data, 0644); err != nil {
		return err
	}

	log.Infof("file %v has been successfully written", c.File)
	return nil
}

// getDriversReleases retrieves DriversReleases using discovery and github API
func getDriversReleases(ctx context.Context) (DriversReleases, error) {
	log.Infof("getting the list of supported drivers")
	drivers, err := discovery.OfficialDrivers(ctx, &discovery.Options{
		NoBuildInfo:   true,
		NoMaintainers: true,
		NoSDKVersion:  true,
		NoStatic:      true,
	})
	if err != nil {
		return nil, errFailedToGetDriversReleases.Wrap(err)
	}

	results := make(DriversReleases)
	for _, d := range drivers {
		version, err := d.LatestVersion(ctx)
		if err != nil {
			return nil, errFailedToGetDriversReleases.Wrap(err)
		}

		var (
			versionString = version.String()
			tag           = "v" + versionString
			repo          = d.Language + "-driver"
		)

		log.Infof("driver: %v, version: %v", d.Language, version.String())
		if versionString == "0.0.0" {
			log.Infof("skipping empty version for %v", repo)
			continue
		}

		description, err := getCommitsDescription(ctx, repo, tag)
		if err != nil {
			return nil, errFailedToGetDriversReleases.Wrap(err)
		}
		if description == "" {
			log.Infof("%v: no additional commits detected", repo)
			continue
		}

		log.Infof("post-commits: %v", description)
		results[d.Language] = Release{
			Tag:         tag,
			Description: description,
		}
	}

	return results, nil
}

// TODO(lwsanty): we probably need to adjust commits filtering by branch, but it's not crucial while we work on fork PRs basis
// getCommitsDescription is used to obtain concatenated descriptions of commits, performed after release
func getCommitsDescription(ctx context.Context, repo, tag string) (string, error) {
	client := discovery.GithubClient()
	release, _, err := client.Repositories.GetReleaseByTag(ctx, owner, repo, tag)
	if err != nil {
		return "", errFailedToGetCommitsDescription.New(repo, err)
	}

	var results []string
	for page := 1; ; page++ {
		rCommits, _, err := client.Repositories.ListCommits(ctx, owner, repo, &github.CommitsListOptions{
			Since: release.PublishedAt.UTC(),
			ListOptions: github.ListOptions{
				Page: page, PerPage: 100,
			},
		})
		if err != nil {
			return "", errFailedToGetCommitsDescription.New(repo, err)
		}

		log.Debugf("commits: %v", len(rCommits))
		if len(rCommits) == 0 {
			break
		}
		for _, c := range rCommits {
			log.Debugf("appending: %v", *c.Commit.Message)
			results = append(results, processCommitMessage(*c.Commit.Message))
		}
	}

	return strings.Join(results, "\n"), nil
}

// processCommitMessage performs required formatting to represent concatenated commits as a bullet list
func processCommitMessage(msg string) string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(msg))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Signed-off") || line == "" {
			continue
		}
		lines = append(lines, line)
	}

	return "* " + strings.Join(lines, "\n")
}
