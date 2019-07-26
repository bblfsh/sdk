package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"

	"github.com/bblfsh/sdk/v3/cmd/bblfsh-drivers-updater/utils"
	"github.com/bblfsh/sdk/v3/driver/manifest/discovery"

	"gopkg.in/src-d/go-log.v1"
)

func main() {
	branchPtr := flag.String("branch", "patch-1", "branch to be created")
	SDKVersionPtr := flag.String("sdk-version", "", "sdk version to update to")
	scriptPathPtr := flag.String("script", "", "path to the script that will be executed")
	dockerfilePtr := flag.Bool("dockerfile", false, "use dockerfile to create a branch")
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
		if *SDKVersionPtr == "" {
			log.Infof("both script and SDK version not found, exiting")
			os.Exit(0)
		}
		fallthrough
	default:
		scriptText = string(scriptData)
	}

	drivers, err := discovery.OfficialDrivers(context.Background(), &discovery.Options{
		NoStatic:      true,
		NoMaintainers: true,
		NoBuildInfo:   true,
	})
	handleErr(err)

	for _, d := range drivers {
		log.Infof("l: %+v, a: %+v, v: %+v", d.Language, d.RepositoryURL(), d.SDKVersion)
		tmpSDKVersion := *SDKVersionPtr
		switch {
		case d.InDevelopment():
			log.Infof("skipping driver %v: not supported or still in development", d.Language)
			continue
		case tmpSDKVersion == d.SDKVersion:
			log.Infof("driver %v: sdk %v is already installed", d.Language, tmpSDKVersion)
			tmpSDKVersion = ""
			if scriptText == "" {
				log.Infof("skipping driver %v: script is empty and version update is not required", d.Language)
				continue
			}
			fallthrough
		default:
			handleErr(utils.PrepareBranch(d, &utils.UpdateOptions{
				Branch:     *branchPtr,
				SDKVersion: tmpSDKVersion,
				Script:     scriptText,
				Dockerfile: *dockerfilePtr,
			}))
			handleErr(utils.PreparePR(d, *branchPtr))
		}
	}
}
