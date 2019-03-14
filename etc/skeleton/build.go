package main

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/bblfsh/sdk.v2/build"
)

func main() {
	flag.Parse()
	if err := runBuild("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runBuild(root string) error {
	args := flag.Args()
	name := ""
	if len(args) != 0 {
		name = args[0]
	}
	d, err := build.NewDriver(root)
	if err != nil {
		return err
	}
	id, err := d.Build(name)
	if err != nil {
		return err
	}
	fmt.Println(id)
	return nil
}
