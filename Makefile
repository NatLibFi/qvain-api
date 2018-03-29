#
# -wvh- Makefile to build Go binaries for commands in cmd/*
#
#       The only thing this Makfile does as opposed to an ordinary `go build ./...`

GO := go
CMDS := $(notdir $(wildcard cmd/*))
BINDIR := $(PWD)/bin
DATADIRS := $(addprefix $(PWD)/,doc bench bin)
SOURCELINK := ${GOBIN}/sourcelink

# VCS
TAG := $(shell git describe --always 2>/dev/null)
HASH := $(shell git rev-parse --short HEAD 2>/dev/null)
BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
REPO := $(shell git ls-remote --get-url 2>/dev/null)
#REPOLINK := $(shell test -x $(SOURCELINK) && ${GOBIN}/sourcelink $(REPO) $(HASH) $(BRANCH) 2>/dev/null || echo)
VERSION_PACKAGE := $(shell $(GO) list -f '{{.ImportPath}}' ./version)

#IMPORT_PATH := $(shell go list -f '{{.ImportPath}}' .)
#BINARY := $(notdir $(IMPORT_PATH))

.PHONY: all install run runall release clean cloc doc prebuild listall

all: listall $(CMDS)
	@echo built all: $(CMDS)

$(CMDS): prebuild $(wildcard cmd/$@/*.go)
	@echo building: $@
	@cd cmd/$@; \
	$(GO) build -o $(BINDIR)/$@ -ldflags "-s -w -X $(VERSION_PACKAGE).CommitHash=$(HASH) -X $(VERSION_PACKAGE).CommitTag=$(TAG) -X $(VERSION_PACKAGE).CommitBranch=$(BRANCH) -X $(VERSION_PACKAGE).CommitRepo=$(REPOLINK)"

install: all
	@echo $(DATADIRS)

# hack to run command from make command line goal arguments
# NOTE: any clean-up lines after the command is run won't execute if the program is interrupted with SIGINT
.SECONDEXPANSION:
runall: $$(filter-out $$@,$(MAKECMDGOALS))
	@- bash -c "trap 'true' SIGINT; $(BINDIR)/$<" || rm -f $(BINDIR)/$<
	rm -f $(addprefix $(BINDIR)/, $^)

# hack to run command from make command line goal arguments
# Supports simple arguments but won't work for complex arguments because Make splits on spaces.
# Remember to escape flags so Make doesn't interpret them:
#   $ make -- run some-command -d
# NOTE: any clean-up lines after the command is run won't execute if the program is interrupted with SIGINT
.SECONDEXPANSION:
run: $$(wordlist 2,2,$(MAKECMDGOALS))
	@- bash -c "trap 'true' SIGINT; $(BINDIR)/$< $(wordlist 3,100,$(MAKECMDGOALS))" || rm -f $(BINDIR)/$<
	rm -f $(addprefix $(BINDIR)/, $^)

clean:
	#rm -f $(foreach cmd,$(CMDS),cmd/$(cmd)/$(cmd))
	go clean ./...
	rm -f $(BINDIR)/*

doc: doc/go_dependencies.md
	@cd doc && ./make_go_dependencies_list.sh

$(SOURCELINK):
	-go get -v wvh/sourcelink

prebuild: $(SOURCELINK)
	@$(eval REPOLINK=$(shell test -x ${GOBIN}/sourcelink && ${GOBIN}/sourcelink $(REPO) $(HASH) $(BRANCH) 2>/dev/null || echo ""))
	@echo ran prebuild requirements

release: all doc

cloc:
	cloc --exclude-dir=vendor .

listall:
	@echo version: $(TAG)
	@echo building all: $(CMDS)
