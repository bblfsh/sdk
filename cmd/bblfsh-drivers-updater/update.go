package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bblfsh/sdk/v3/cmd/bblfsh-drivers-updater/utils"
	"github.com/bblfsh/sdk/v3/driver/manifest/discovery"

	"gopkg.in/src-d/go-log.v1"
)

const (
	commitMsg = "autogenerated changes"
)

func main() {
	branchPtr := flag.String("branch", "patch-1", "branch to be created")
	SDKVersionPtr := flag.String("sdk-version", "", "sdk version to update to")
	scriptPathPtr := flag.String("script", "", "path to the script that will be executed")
	commitMsgPtr := flag.String("commit-msg", commitMsg, "commit message of the update")
	dockerfilePtr := flag.Bool("dockerfile", false, "use dockerfile to create a branch")
	dryRunPtr := flag.Bool("dry-run", false, "dry run")
	flag.Parse()

	handleErr := func(err error) {
		if err != nil {
			log.Errorf(err, "error")
			os.Exit(1)
		}
	}

	var scriptText string
	scriptData, err := ioutil.ReadFile(*scriptPathPtr)

	switch {
	case err != nil && !os.IsNotExist(err):
		log.Errorf(err, "error")
		os.Exit(1)
	case os.IsNotExist(err):
		log.Infof("script %v does not exist", *scriptPathPtr)
		if strings.TrimSpace(*SDKVersionPtr) == "" {
			log.Infof("both script and SDK version not found, exiting")
			os.Exit(0)
		}
		fallthrough
	case strings.TrimSpace(string(scriptData)) == "" && strings.TrimSpace(*SDKVersionPtr) == "":
		log.Infof("script and SDK version are empty, nothing to do here")
		os.Exit(0)
	default:
		scriptText = string(scriptData)
	}

	log.Infof("getting the list of supported drivers")
	drivers, err := discovery.OfficialDrivers(context.Background(), &discovery.Options{
		NoStatic:      true,
		NoMaintainers: true,
		NoBuildInfo:   true,
	})
	handleErr(err)

	log.Infof("%v drivers found", len(drivers))
	for _, d := range drivers {
		log.Infof("Processing driver language: %+v, URL: %+v, SDK version: %+v", d.Language, d.RepositoryURL(), d.SDKVersion)
		tmpSDKVersion := *SDKVersionPtr
		switch {
		case d.InDevelopment():
			log.Infof("skipping driver %v: not supported or still in development", d.Language)
			continue
		case tmpSDKVersion == d.SDKVersion:
			log.Infof("driver %v: sdk %v is already installed", d.Language, tmpSDKVersion)
			tmpSDKVersion = ""
			if strings.TrimSpace(scriptText) == "" {
				log.Infof("skipping driver %v: script is empty and version update is not required", d.Language)
				continue
			}
			fallthrough
		default:
			githubToken := os.Getenv("GITHUB_TOKEN")
			err := utils.PrepareBranch(d, githubToken, &utils.UpdateOptions{
				Branch:     *branchPtr,
				SDKVersion: tmpSDKVersion,
				Script:     scriptText,
				CommitMsg:  *commitMsgPtr,
				Dockerfile: *dockerfilePtr,
				DryRun:     *dryRunPtr,
			})
			if utils.ErrNothingToCommit.Is(err) {
				log.Warningf("skipping driver %s: nothing to change", d.Language)
				continue
			}
			handleErr(err)
			handleErr(utils.PreparePR(d, githubToken, *branchPtr, *commitMsgPtr, *dryRunPtr))
		}
	}
}
