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
// are also documented if you run `./uast-diff-create-testdata --help`.
// $ ./uast-diff-create-testdata -d ~/data/sourced/treediff/python-dataset -f smalltest.txt -o . | sh -

var datasetPath = flag.String("d", "./", "Path to the python-dataset (unpacked flask.tar.gz)")
var testnamesPath = flag.String("f", "./smalltest.txt", "File with testnames")
var outPath = flag.String("o", "./", "Output directory")

func firstFile(dirname string, fnamePattern string) string {
	fns, err := filepath.Glob(filepath.Join(dirname, fnamePattern))
	if err != nil {
		panic(err)
	}
	return fns[0]
}

func main() {
	flag.Parse()
	_, err := os.Open(*datasetPath)
	if err != nil {
		panic(err)
	}

	testnames, err := os.Open(*testnamesPath)
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(testnames)
	scanner.Split(bufio.ScanLines)

	i := 0
	for scanner.Scan() {
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
		i = i + 1
	}

}
