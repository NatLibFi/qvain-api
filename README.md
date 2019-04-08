[![GoDoc](https://godoc.org/github.com/CSCfi/qvain-api?status.svg)](https://godoc.org/github.com/CSCfi/qvain-api)
<!--
[![Build Status](https://travis-ci.org/NatLibFi/qvain-api.svg?branch=next)](https://travis-ci.org/NatLibFi/qvain-api)
-->
[![Go Report Card](https://goreportcard.com/badge/github.com/CSCfi/qvain-api)](https://goreportcard.com/report/github.com/CSCfi/qvain-api)

# Qvain API

## Introduction

Qvain is a web application for description of metadata based on a [JSON schema](http://json-schema.org/). It is developed by the [National Library of Finland](https://www.kansalliskirjasto.fi/en) as part of the [FairData](https://fairdata.fi/) project.

This repository contains the server backend, written in the [Go](https://golang.org/) programming language.

If you are looking for technical information, the `doc/` directory in the source contains a [list of Qvain related links](doc/links.md) as well as a [brief introduction](doc/using_go.md) to setting up Go for those who would like to contribute.

## Installation

### Install Go

The Go project releases a new version every half year, and only supports the last two releases. The best version for compiling this application is the latest available from the [official Go website](https://golang.org/).

Before Go 1.11, the language did not have built-in dependency management; `go get` always downloaded the latest version available from dependency repositories. More recent versions of Go have an official dependency tool called [Go Modules](https://github.com/golang/go/wiki/Modules). It's strongly suggested to use a recent version of Go that has support for modules so dependency issues are avoided and compilation would give you the exact same end result.


### Get code

If you run a recent version of Go, just clone the repository wherever you want. The build process will automatically install the correct dependency versions from the lock file; you don't need to do anything.

```shell
$ git clone https://github.com/CSCfi/qvain-api
$ cd qvain-api
```

If you can't update to a recent version of Go, you'll have to make sure the source is checked out in your `GOPATH`, and check out the dependencies manually with `go get`:

```shell
$ go env GOPATH
/home/jack/GoPath
$ cd $GOPATH
$ mkdir -p src/github.com/CSCfi
$ cd src/github.com/CSCfi
$ git clone https://github.com/CSCfi/qvain-api
$ cd qvain-api
$ go get -v ./cmd/...
```

Note: The triple dot syntax `...` in Go commands means _"and anything below that"_. So that last command should get all dependencies for anything in the `cmd/` directory.

### Build

You can build this application with the included Makefile or with standard Go commands. The benefit of the Makefile is that it will insert version information during the compilation; prefer this for "official" releases running on real servers.

```shell
$ make all
```

Compiled binaries will end up in the `bin/` directory.


If you don't have `make` installed, you can build the application with standard Go commands:

```shell
$ GOBIN=$PWD/bin go install -v ./cmd/...
```

... This will build the commands in the `cmd/` directory and install them into the `bin/` directory. The `GOBIN` environment variable points to the location Go will store compiled binaries.

It's preferred to store the binaries into `bin/` so they don't get accidentally checked into the source repository.


### Rebuild

To re-compile the code after modification, just run `make all` or `go install` like above.

If you only want to (re)build one program – let's say the backend – then just use `go build` directly:

```shell
$ go build -o bin/qvain-backend ./cmd/qvain-backend
```

This will build the source for the backend from the directory `./cmd/qvain-backend` and store/overwrite the executable in `bin/`.

Alternatively, you can change into the directory of the command you want to build and just run `go build`:

```shell
$ cd cmd/qvain-backend
$ go build -v
```

Note that in this case you will have to run `go clean` because the binary will sit in the source directory and we don't want to check it into the code repository.

### Clean

Run `make clean` or just delete the contents of `bin/`:

```shell
$ rm bin/*
```

If you have built binaries inside the source tree instead of outputting them to `bin/`, run `go clean`:

```shell
$ go clean ./...
```

This will find all build artifacts in the source code and delete them. It will not touch `bin/` because that's made by us.

### Run

Go compiles to binary executable files, so just run the command:

```shell
$ bin/metax-cli
metax-query (version 178e2ee)
usage: bin/metax-cli <sub-command> [flags]

  datasets   query dataset API endpoint
  fetch      fetch from dataset API endpoint
  publish    publish dataset to API endpoint
  version    query version

```
