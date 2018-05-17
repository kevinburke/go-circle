# go-circle

[![CircleCI](https://circleci.com/gh/kevinburke/go-circle.svg?style=svg)](https://circleci.com/gh/kevinburke/go-circle)

This is a wrapper for the CircleCI API. Currently we use it to fetch the latest
build for a branch, and wait for tests to finish on that branch. This fork
exists because I left Shyp and wanted to keep developing the tool.

## Usage

```
$ circle
The circle binary interacts with a server that runs your tests.

Usage:

	circle command [arguments]

The commands are:

	cancel              Cancel the current build.
	enable              Enable CircleCI tests for this project.
	open                Open the latest branch build in a browser.
	rebuild             Rebuild a given test branch.
	version             Print the current version
	wait                Wait for tests to finish on a branch.
	download-artifacts  Download all artifacts.

Use "circle help [command]" for more information about a command.
```

In particular, `circle wait` will print live statistics about how long each of
your build steps are taking in each container. If the build fails, `circle wait`
will download the console output from the failed build step, and display it in
the console. `wait` also displays statistics about how long each step of your
build took.

```
$ circle wait
Waiting for latest build on my-branch to complete
Build on my-branch succeeded!

Step                                         0
=====================================================
Spin up Environment                          1.07s
Checkout code                                730ms
Restoring Cache                              250ms
make race-test                               16.76s
Saving Cache                                 80ms

Tests on my-branch took 21s. Quitting.
```

## Token Management

This library will look for your Circle API token in `~/cfg/circleci` and (if
that does not exist, in `~/.circlerc`). The configuration file should look like
this:

```toml
[organizations]

    [organizations.kevinburke]
    token = "aabbccddeeff00"
```

You can specify any organization name you want.

## Installation

Find your target operating system (darwin, windows, linux) and desired bin
directory, and modify the command below as appropriate:

    curl --silent --location --output=/usr/local/bin/circle https://github.com/kevinburke/circle/releases/download/0.33/circle-linux-amd64 && chmod 755 /usr/local/bin/circle

The latest version is 0.33.

If you have a Go development environment, you can also install via source code:

```
go get -u github.com/kevinburke/go-circle/...
```

This should place a `circle` binary in `$GOPATH/bin`, so for me,
`~/bin/circle`.
