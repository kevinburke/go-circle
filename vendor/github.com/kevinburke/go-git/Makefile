.PHONY: test

BENCHSTAT := $(GOPATH)/bin/benchstat
BUMP_VERSION := $(GOPATH)/bin/bump_version
GODOCDOC := $(GOPATH)/bin/godocdoc
MEGACHECK := $(GOPATH)/bin/megacheck
UNAME := $(shell uname -s)

$(GOPATH)/bin:
	mkdir -p $(GOPATH)/bin

$(MEGACHECK): $(GOPATH)/bin
ifeq ($(UNAME), Darwin)
	curl --silent --location --output $(MEGACHECK) https://github.com/kevinburke/go-tools/releases/download/2018-04-15/megacheck-darwin-amd64
endif
ifeq ($(UNAME), Linux)
	curl --silent --location --output $(MEGACHECK) https://github.com/kevinburke/go-tools/releases/download/2018-04-15/megacheck-linux-amd64
endif
	chmod 755 $(MEGACHECK)

lint: $(MEGACHECK)
	go list ./... | grep -v vendor | xargs $(MEGACHECK) --ignore='github.com/kevinburke/go-git/*.go:U1000'
	go list ./... | grep -v vendor | xargs go vet

$(GODOCDOC):
	go get github.com/kevinburke/godocdoc

docs: $(GODOCDOC)
	$(GODOCDOC)

test: lint
	@# this target should always be listed first so "make" runs the tests.
	go test ./...

race-test: lint
	go test -race ./...

release: | $(BUMP_VERSION)
	$(BUMP_VERSION) minor git.go
