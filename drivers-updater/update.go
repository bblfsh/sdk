package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bblfsh/sdk/v3/drivers-updater/utils"
	"gopkg.in/src-d/go-log.v1"
)

func main() {
	branchPtr := flag.String("branch", "patch-1", "branch to be created")
	// TODO(lwsanty): file option maybe
	scriptPtr := flag.String("script", "", "script to be executed")
	flag.Parse()

	res, err := utils.GetDrivers()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Infof("%+v", res[3])

	for _, r := range res {
		if err := utils.PrepareBranch(r, *branchPtr, *scriptPtr); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := utils.PreparePR(r, *branchPtr); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
