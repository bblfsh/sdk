package main

import (
	"context"
	"flag"
	"io/ioutil"
	"os"

	"github.com/bblfsh/sdk/v3/cmd/drivers-updater/utils"
	"github.com/bblfsh/sdk/v3/driver/manifest/discovery"
	"gopkg.in/src-d/go-log.v1"
)

func main() {
	branchPtr := flag.String("branch", "patch-1", "branch to be created")
	SDKVersionPtr := flag.String("sdk-version", "", "sdk version to update to")
	scriptPathPtr := flag.String("script", "", "path to the script that will be executed")
	flag.Parse()

	handleErr := func(err error) {
		if err != nil {
			log.Errorf(err, "error")
			os.Exit(1)
		}
	}

	scriptData, err := ioutil.ReadFile(*scriptPathPtr)
	handleErr(err)

	drivers, err := discovery.OfficialDrivers(context.Background(), nil)
	handleErr(err)

	for _, d := range drivers {
		log.Infof("l: %+v, a: %+v, v: %+v", d.Language, d.RepositoryURL(), d.SDKVersion)
		tmpSDKVersion := *SDKVersionPtr
		switch {
		case d.SDKVersion == "1":
			log.Infof("skipping driver %v: not supported", d.Language)
			continue
		case tmpSDKVersion == d.SDKVersion:
			log.Infof("driver %v: sdk %v is already installed", d.Language, tmpSDKVersion)
			tmpSDKVersion = ""
			if string(scriptData) == "" {
				log.Infof("skipping driver %v: script is empty and version update is not required", d.Language)
				continue
			}
			fallthrough
		default:
			handleErr(utils.PrepareBranch(d, *branchPtr, tmpSDKVersion, string(scriptData)))
			handleErr(utils.PreparePR(d, *branchPtr))
		}
	}
}
