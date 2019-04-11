package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// This is a script used to generate test data for the diff library. It uses pairs of files
// to get diffed acquired from
// https://github.com/vmarkovtsev/treediff/blob/49356e7f85c261ed88cf46326791765c58c22b5b/dataset/flask.tar.xz
// It uses https://github.com/bblfsh/client-go#Installation to convert python sources into
// uast yaml files.
// This program outputs commands on the stdout - they are to be piped to sh/bash. It's a good
// idea to inspect them before running by just reading the textual output of the program.
// The program needs to be ran with proper commandline arguments. Example below. The cli arguments
// are also documented if you run `go run ./uast-diff-create-testdata --help`.
// $ go run ./create-testdata/ -d ~/data/sourced/treediff/python-dataset -f smalltest.txt -o . | sh -

var (
	datasetPath   = flag.String("d", "./", "Path to the python-dataset (unpacked flask.tar.gz)")
	testnamesPath = flag.String("f", "./smalltest.txt", "File with testnames")
	outPath       = flag.String("o", "./", "Output directory")
)

func firstFile(dirname string, fnamePattern string) string {
	fns, err := filepath.Glob(filepath.Join(dirname, fnamePattern))
	if err != nil {
		panic(err)
	}
	return fns[0]
}

func main() {
	flag.Parse()
	_, err := os.Stat(*datasetPath)
	if err != nil {
		panic(err)
	}

	testnames, err := os.Open(*testnamesPath)
	if err != nil {
		panic(err)
	}
	defer testnames.Close()

	scanner := bufio.NewScanner(testnames)

	for i := 0; scanner.Scan(); i++ {
		name := scanner.Text()
		if name == "" {
			continue
		}
		src := firstFile(*datasetPath, name+"_before*.src")
		dst := firstFile(*datasetPath, name+"_after*.src")
		iStr := strconv.Itoa(i)
		fmt.Println("bblfsh-cli -l python " + src + " -o yaml > " +
			filepath.Join(*outPath, iStr+"_src.uast"))
		fmt.Println("bblfsh-cli -l python " + dst + " -o yaml > " +
			filepath.Join(*outPath, iStr+"_dst.uast"))
	}
}
