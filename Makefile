#
# -wvh- Makefile to build Go binaries for commands in cmd/*
#
#       The only thing this Makfile does as opposed to an ordinary `go build ./...`

GO := go
CMDS := $(notdir $(wildcard cmd/*))
BINDIR := $(PWD)/bin
DATADIRS := $(addprefix $(PWD)/,doc bench bin)
TAG := $(shell git describe --always)
HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
VERSION_PACKAGE := $(shell $(GO) list -f '{{.ImportPath}}' ./version)

#IMPORT_PATH := $(shell go list -f '{{.ImportPath}}' .)
#BINARY := $(notdir $(IMPORT_PATH))

.PHONY: all install release clean cloc doc listall

all: listall $(CMDS)
	@echo built all: $(CMDS)

$(CMDS): $(wildcard cmd/$@/*.go)
	@echo building: $@
	@cd cmd/$@; \
	$(GO) build -o $(BINDIR)/$@ -ldflags "-s -w -X $(VERSION_PACKAGE).CommitHash=$(HASH) -X $(VERSION_PACKAGE).CommitTag=$(TAG)"

install: all
	@echo $(DATADIRS)

clean:
	#rm -f $(foreach cmd,$(CMDS),cmd/$(cmd)/$(cmd))
	go clean ./...
	rm -f $(BINDIR)/*

doc: doc/go_dependencies.md
	@cd doc && ./make_go_dependencies_list.sh

release: all doc

cloc:
	cloc --exclude-dir=vendor .

listall:
	@echo version: $(TAG)
	@echo building all: $(CMDS)
