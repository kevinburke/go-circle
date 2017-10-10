.PHONY: install test

STATICCHECK := $(GOPATH)/bin/staticcheck
BUMP_VERSION := $(GOPATH)/bin/bump_version

install:
	go install ./...

build:
	go get ./...
	go build ./...

$(STATICCHECK):
	go get -u honnef.co/go/tools/cmd/staticcheck

lint: $(STATICCHECK)
	go vet ./...
	$(STATICCHECK) ./...

test: install lint
	go test -v -race ./...

$(BUMP_VERSION):
	go get github.com/Shyp/bump_version

release: $(BUMP_VERSION)
	git checkout master
	$(BUMP_VERSION) minor circle.go
	git push origin master
	git push origin master --tags
