.PHONY: install test

MEGACHECK := $(GOPATH)/bin/megacheck
BUMP_VERSION := $(GOPATH)/bin/bump_version
WRITE_MAILMAP := $(GOPATH)/bin/write_mailmap

install:
	go install ./...

build:
	go get ./...
	go build ./...

$(MEGACHECK):
	go get -u honnef.co/go/tools/cmd/megacheck

lint: $(MEGACHECK)
	go vet ./...
	go list ./... | grep -v vendor | xargs $(MEGACHECK) --ignore='github.com/kevinburke/go-circle/*.go:S1002'

test: install lint
	go test -v -race ./...

$(BUMP_VERSION):
	go get github.com/Shyp/bump_version

release: $(BUMP_VERSION)
	git checkout master
	$(BUMP_VERSION) minor circle.go
	git push origin master
	git push origin master --tags

$(WRITE_MAILMAP):
	go get -u github.com/kevinburke/write_mailmap

AUTHORS.txt: | $(WRITE_MAILMAP)
	$(WRITE_MAILMAP) > AUTHORS.txt

authors: AUTHORS.txt
	write_mailmap > AUTHORS.txt
