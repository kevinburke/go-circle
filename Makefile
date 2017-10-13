.PHONY: install test

MEGACHECK := $(GOPATH)/bin/megacheck
BUMP_VERSION := $(GOPATH)/bin/bump_version
WRITE_MAILMAP := $(GOPATH)/bin/write_mailmap
RELEASE := $(GOPATH)/bin/github-release

build:
	go get ./...
	go build ./...

$(MEGACHECK):
	go get -u honnef.co/go/tools/cmd/megacheck

lint: $(MEGACHECK)
	go vet ./...
	go list ./... | grep -v vendor | xargs $(MEGACHECK) --ignore='github.com/kevinburke/go-circle/*.go:S1002'

test: lint
	go test -v -race ./...

$(BUMP_VERSION):
	go get -u github.com/Shyp/bump_version

$(RELEASE):
	go get -u github.com/aktau/github-release

release: test | $(BUMP_VERSION) $(RELEASE)
ifndef version
	@echo "Please provide a version"
	exit 1
endif
ifndef GITHUB_TOKEN
	@echo "Please set GITHUB_TOKEN in the environment"
	exit 1
endif
	$(BUMP_VERSION) --version=$(version) circle.go
	git push origin --tags
	mkdir -p releases/$(version)
	GOOS=linux GOARCH=amd64 go build -o releases/$(version)/circle-linux-amd64 ./circle
	GOOS=darwin GOARCH=amd64 go build -o releases/$(version)/circle-darwin-amd64 ./circle
	GOOS=windows GOARCH=amd64 go build -o releases/$(version)/circle-windows-amd64 ./circle
	# These commands are not idempotent, so ignore failures if an upload repeats
	$(RELEASE) release --user kevinburke --repo go-circle --tag $(version) || true
	$(RELEASE) upload --user kevinburke --repo go-circle --tag $(version) --name circle-linux-amd64 --file releases/$(version)/circle-linux-amd64 || true
	$(RELEASE) upload --user kevinburke --repo go-circle --tag $(version) --name circle-darwin-amd64 --file releases/$(version)/circle-darwin-amd64 || true
	$(RELEASE) upload --user kevinburke --repo go-circle --tag $(version) --name circle-windows-amd64 --file releases/$(version)/circle-windows-amd64 || true

$(WRITE_MAILMAP):
	go get -u github.com/kevinburke/write_mailmap

AUTHORS.txt: | $(WRITE_MAILMAP)
	$(WRITE_MAILMAP) > AUTHORS.txt

authors: AUTHORS.txt
	write_mailmap > AUTHORS.txt
