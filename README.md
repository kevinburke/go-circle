# go-circle

[![CircleCI](https://circleci.com/gh/kevinburke/go-circle.svg?style=svg)](https://circleci.com/gh/kevinburke/go-circle)

This is a very incomplete wrapper for the CircleCI API. Currently we use it to
fetch the latest build for a branch.

You should treat the API as very unstable, library API's that grow from one or
two methods to the whole API tend to not be designed very well, so probably at
some point you will have to create a Client instance or something.

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

If you want to install the project, first set your `$GOPATH` in your
environment (I set it to `$HOME`), then run

```
go install github.com/kevinburke/go-circle/...
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
