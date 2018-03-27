.PHONY: install build test
BUMP_VERSION := $(GOPATH)/bin/bump_version
GODOCDOC := $(GOPATH)/bin/godocdoc
MEGACHECK := $(GOPATH)/bin/megacheck

install:
	go get ./...
	go install ./...

build:
	bazel build //...

$(MEGACHECK):
	go get honnef.co/go/tools/cmd/megacheck

vet: $(MEGACHECK)
	$(MEGACHECK) ./...
	go vet ./...

test: vet
	bazel test --test_output=errors //...

race-test:
	bazel test --test_output=errors --features=race //...

ci:
	bazel --batch test \
		--noshow_progress --noshow_loading_progress \
		--test_output=errors \
		--features=race //...

$(BUMP_VERSION):
	go get github.com/Shyp/bump_version

release: test | $(BUMP_VERSION)
	$(BUMP_VERSION) minor types.go

docs: $(GODOCDOC)
	$(GODOCDOC)
