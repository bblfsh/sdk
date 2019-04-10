package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bblfsh/sdk/v3/build"
)

var (
	fBblfshd = flag.String("bblfshd", "", "bblfshd version to test with")
	fBench   = flag.Bool("bench", false, "benchmark the driver")
)

func main() {
	flag.Parse()
	if err := runTest("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runTest(root string) error {
	args := flag.Args()
	image := ""
	if len(args) != 0 {
		image = args[0]
	}
	d, err := build.NewDriver(root)
	if err != nil {
		return err
	}
	return d.Test(*fBblfshd, image, *fBench)
}
