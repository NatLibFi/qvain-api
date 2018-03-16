GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v
GOGET=$(GOCMD) get -v

IMPORT_PATH := $(shell go list -f '{{.ImportPath}}' .)
BINARY := $(notdir $(IMPORT_PATH))

TAG := $(shell git describe --always)
HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION_PACKAGE := $(shell go list -f '{{.ImportPath}}' ../../version)

.PHONY: all
all: build

build:
	#@echo GOPATH=$(value GOPATH)
	#@echo GOBIN=$(value GOBIN)
	$(GOBUILD) -ldflags "-s -w -X $(VERSION_PACKAGE).CommitHash=$(HASH) -X $(VERSION_PACKAGE).CommitTag=$(TAG)"

run:
	-$(GORUN) -ldflags "-s -w -X $(VERSION_PACKAGE).CommitHash=$(HASH) -X $(VERSION_PACKAGE).CommitTag=$(TAG)" *.go

clean:
	rm -f $(BINARY)

update:
	git pull
	#git submodule update --init
