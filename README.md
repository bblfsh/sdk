# sdk [![Build Status](https://travis-ci.org/bblfsh/sdk.svg?branch=master)](https://travis-ci.org/bblfsh/sdk) [![codecov](https://codecov.io/gh/bblfsh/sdk/branch/master/graph/badge.svg)](https://codecov.io/gh/bblfsh/sdk) [![license](https://img.shields.io/badge/license-GPL--3.0-blue.svg)](https://github.com/bblfsh/sdk/blob/master/LICENSE) [![GitHub release](https://img.shields.io/github/release/bblfsh/sdk.svg)](https://github.com/bblfsh/sdk/releases)

Babelfish SDK contains the tools and libraries
required to create a Babelfish driver for a programming language.

## Build

### Dependencies

The Babelfish SDK has the following dependencies:

* [Docker](https://www.docker.com/get-docker)
* [Go](https://golang.org/dl/)

Make sure that you've correctly set your [GOROOT and
GOPATH](https://golang.org/doc/code.html#Workspaces) environment variables.

### Install

Babelfish SDK gets installed using either Go:

```bash
$ go get -t -v gopkg.in/bblfsh/sdk.v2/...
```

or make command:

```bash
$ make install
```

These commands will install `bblfsh-sdk` program at `$GOPATH/bin/`.

### Contribute

The SDK provides scaffolding templates for creating a new language driver.
These templates are converted to Go code that ends up in `bblfsh-sdk` tool. Use `make` to update these templates:

```bash
$ make
go get -v github.com/jteeuwen/go-bindata/...
go get -v golang.org/x/tools/cmd/cover/...
cat protocol/internal/testdriver/main.go | sed -e 's|\([[:space:]]\+\).*//REPLACE:\(.*\)|\1\2|g' \
	> etc/skeleton/driver/main.go.tpl
chmod -R go=r ${GOPATH}/src/github.com/bblfsh/sdk/etc/build; \
go-bindata \
	-pkg build \
	-modtime 1 \
	-nocompress \
	-prefix ${GOPATH}/src/github.com/bblfsh/sdk/etc/build \
	-o assets/build/bindata.go \
	${GOPATH}/src/github.com/bblfsh/sdk/etc/build/...
chmod -R go=r ${GOPATH}/src/github.com/bblfsh/sdk/etc/skeleton; \
go-bindata \
	-pkg skeleton \
	-modtime 1 \
	-nocompress \
	-prefix ${GOPATH}/src/github.com/bblfsh/sdk/etc/skeleton \
	-o assets/skeleton/bindata.go \
	${GOPATH}/src/github.com/bblfsh/sdk/etc/skeleton/...
```

You can validate this process has been properly done before submitting changes:

```bash
$ make validate-commit
```

If the code has not been properly generated,
this command will show a diff of the changes that have not been processed
and will end up with a message like:

```
generated bindata is out of sync
make: *** [Makefile:66: validate-commit] Error 2
```

Review the process if this happens.

On the other hand, If you need to regenerate *[proto](https://developers.google.com/protocol-buffers/)*  and *[proteus](https://github.com/src-d/proteus)* files, you must run `go generate` from *protocol/* directory:

```bash
$ cd protocol/
$ go generate
```

It regenerates all *[proto](https://developers.google.com/protocol-buffers/)* and *[proteus](https://github.com/src-d/proteus)* files under *[protocol/](https://github.com/bblfsh/sdk/tree/master/protocol)* and *[uast/](https://github.com/bblfsh/sdk/tree/master/uast)* directories.

## Usage

Babelfish SDK helps both setting up the initial structure of a new driver
and keeping that structure up to date.

### Creating the driver's initial structure

Let's say we're creating a driver for `mylang`. The first step is going to the location
where we want the repository for the driver to be bootstrapped:

```bash
$ cd $GOPATH/src/github.com/bblfsh
```

Now the driver should be bootstrapped with `bblfsh-sdk`. This will create a git repository,
and some directories and files required by every driver. They will be overwritten if they
exist, like the README.md file in the example below.

```bash
$ bblfsh-sdk init mylang alpine
initializing driver "mylang", creating new manifest
creating file "manifest.toml"
creating file "Makefile"
creating file "driver/main.go"
creating file "driver/normalizer/normalizer.go"
creating file ".git/hooks/pre-commit"
creating file ".gitignore"
creating file ".travis.yml"
creating file "Dockerfile.build.tpl"
creating file "driver/normalizer/normalizer_test.go"
creating file "Dockerfile.tpl"
creating file "LICENSE"
managed file "README.md" has changed, overriding changes
$ git add -A
$ git commit -m 'initialize repository'
```

Note that this adds a pre-commit [git
hook](https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks), which will verify
these files are up to date before every commit and will disallow commits if some
of the managed files are changed. You can by-pass this with `git commit
--no-verify`.

You can find the driver skeleton used here at [`etc/skeleton`](etc/skeleton).

### Keeping managed files updated

Whenever the managed files are updated, drivers need to update them.
The `update.go` script can be used to perform some of those updates in managed files.
For example, if the README template is updated,
running `go run update.go` will overwrite it.

```bash
$ go run update.go
managed file "README.md" has changed, overriding changes
```

`bblfsh-sdk` doesn't update the SDK itself. Check `Gopkg.toml` for the target SDK version.

For further details of how to construct a language driver,
take a look at [Implementing the driver](https://doc.bblf.sh/driver/sdk.html#implementing-the-driver)
section in documentation.

### Testing the driver

In order to run test for a particular dirver, change to it's directory and run:

```bash
$ go run ./test.go
```

This will:
 - compile a "test binary" that parses content of the `./fixtures` directory of the driver
 - create a docker image with all dependencies, native driver and a test binary
 - run this test binary inside a Docker container, using that image

The test binary first parses all source files in `./fixtures` and generates a set of
`*.native` AST files. Then the second set of test is run that uses `*.native` files,
applies UAST annotations and normalizations to produce `*.uast` and `*.sem.uast` files.

If `*.native` files were already generated, the second stage can be run on the host without
Docker container:
```bash
$ go test ./driver/...
```

Overall, SDK supports 4 different kind of tests for a driver:
 - Native unit tests, parsing source files in `./fixtures` and writing `*.native`. Runs only in Docker.
 - UAST transformation unit tests, using `*.native` and writing `*.uast` files. Can be run both on host and in Docker.
 - Integration tests, using content of `./fixtures/_integration*`. Those are run in Docker using `bblfshd`.
 - Benchmarks, using content of `./fixtures/bench_*.native`. Can be run on the host or in Docker.

First two always run, benchmarks are only triggered by `bblfsh-sdk test --bench` or `go test -bench=. ./driver/...`.

## License

GPLv3, see [LICENSE](LICENSE)

