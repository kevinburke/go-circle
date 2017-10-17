# go-circle

[![CircleCI](https://circleci.com/gh/kevinburke/go-circle.svg?style=svg)](https://circleci.com/gh/kevinburke/go-circle)

This is a wrapper for the CircleCI API. Currently we use it to fetch the latest
build for a branch, and wait for tests to finish on that branch. This fork
exists because I left Shyp and wanted to keep developing the tool.

## Token Management

This library will look for your Circle API token in `~/cfg/circleci` and (if
that does not exist, in `~/.circlerc`). The configuration file should look like
this:

```toml
[organizations]

    [organizations.kevinburke]
    token = "aabbccddeeff00"
```

You can specify any org name you want.

## Installation

Find your target operating system (darwin, windows, linux) and desired bin
directory, and modify the command below as appropriate:

    curl --silent --location --output=/usr/local/bin/circle https://github.com/kevinburke/circle/releases/download/0.22/circle-linux-amd64 && chmod 755 /usr/local/bin/circle

The latest version is 0.22.

If you have a Go development environment, you can also install via source code:

```
go get -u github.com/kevinburke/go-circle/...
```

This should place a `circle` binary in `$GOPATH/bin`, so for me,
`~/bin/circle`.

## Wait for tests to pass/fail on a branch

If you want to be notified when your tests finish running, run `circle wait
[branchname]`. The interface for that will certainly change as well; we should
be able to determine which organization/project to run tests for by checking
your Git remotes.

It's pretty neat! Here's a screenshot.

<img src="https://monosnap.com/file/49h2NvVwxDBtHWlphAGiqzdJFDB7xy.png"
alt="CircleCI screenshot">
