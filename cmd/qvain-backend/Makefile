GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test -v
GOGET=$(GOCMD) get -v

IMPORT_PATH := $(shell go list -f '{{.ImportPath}}' .)
BINARY := $(notdir $(IMPORT_PATH))

TAG := $(shell git describe --always 2>/dev/null)
HASH := $(shell git rev-parse --short HEAD 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
VERSION_PACKAGE := $(shell go list -f '{{.ImportPath}}' ../../version)

.PHONY: all
all: build

build:
	#@echo GOPATH=$(value GOPATH)
	#@echo GOBIN=$(value GOBIN)
	$(GOBUILD) -ldflags "-s -w -X $(VERSION_PACKAGE).CommitHash=$(HASH) -X $(VERSION_PACKAGE).CommitTag=$(TAG) -X $(VERSION_PACKAGE).CommitBranch=$(BRANCH)"

run: build
	#-$(GORUN) -ldflags "-s -w -X $(VERSION_PACKAGE).CommitHash=$(HASH) -X $(VERSION_PACKAGE).CommitTag=$(TAG)" *.go
	./qvain-backend
	rm -f ./qvain-backend

clean:
	rm -f $(BINARY)

update:
	git pull
	#git submodule update --init
