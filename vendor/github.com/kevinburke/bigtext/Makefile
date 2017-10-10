BUMP_VERSION := $(GOPATH)/bin/bump_version
MEGACHECK := $(GOPATH)/bin/megacheck

test:
	go test ./...

race-test:
	go test -race ./...

lint: | $(MEGACHECK)
	go vet ./...
	$(MEGACHECK) ./...

$(MEGACHECK):
	go get honnef.co/go/tools/cmd/megacheck

$(BUMP_VERSION):
	go get github.com/Shyp/bump_version

release: race-test | $(BUMP_VERSION)
	$(BUMP_VERSION) minor bigtext.go
