package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	circle "github.com/kevinburke/go-circle"
	"github.com/kevinburke/go-circle/wait"
	git "github.com/kevinburke/go-git"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/sync/errgroup"
)

const help = `The circle binary interacts with a server that runs your tests.

Usage: 

	circle command [arguments]

The commands are:

	enable              Enable CircleCI tests for this project.
	open                Open the latest branch build in a browser.
	rebuild             Rebuild a given test branch.
	update              Update to the latest version
	version             Print the current version
	wait                Wait for tests to finish on a branch.
	download-artifacts  Download all artifacts.

Use "circle help [command]" for more information about a command.
`

const downloadUsage = `usage: download-artifacts <build-num>`
const enableUsage = `usage: enable [-h]

Turn on CircleCI builds for this project.`

func usage() {
	fmt.Fprintf(os.Stderr, help)
	flag.PrintDefaults()
}

func init() {
	flag.Usage = usage
}

func checkError(err error) {
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

// Given a set of command line args, return the git branch or an error. Returns
// the current git branch if no argument is specified
func getBranchFromArgs(args []string) (string, error) {
	if len(args) == 0 {
		return git.CurrentBranch()
	} else {
		return args[0], nil
	}
}

func doOpen(flags *flag.FlagSet) {
	args := flags.Args()
	branch, err := getBranchFromArgs(args)
	checkError(err)
	remote, err := git.GetRemoteURL("origin")
	checkError(err)
	cr, err := circle.GetTree(remote.Path, remote.RepoName, branch)
	checkError(err)
	if len(*cr) == 0 {
		fmt.Printf("No results, are you sure there are tests for %s/%s?\n",
			remote.Path, remote.RepoName)
		return
	}
	latestBuild := (*cr)[0]
	open.Start(latestBuild.BuildURL)
}

func doDownload(flags *flag.FlagSet) error {
	buildStr := flags.Arg(0)
	val, err := strconv.Atoi(buildStr)
	if err != nil {
		return err
	}
	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}
	arts, err := circle.GetArtifactsForBuild(remote.Path, remote.RepoName, val)
	if err != nil {
		return err
	}
	var g errgroup.Group
	tempDir, err := ioutil.TempDir("", "circle-artifacts")
	if err != nil {
		return err
	}
	for _, art := range arts {
		art := art
		g.Go(func() error {
			return circle.DownloadArtifact(art, tempDir, remote.Path)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote all artifacts for build %d to %s\n", val, tempDir)
	return nil
}

func doEnable(flags *flag.FlagSet) error {
	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return circle.Enable(ctx, remote.Host, remote.Path, remote.RepoName)
}

func doRebuild(flags *flag.FlagSet) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	args := flags.Args()
	branch, err := getBranchFromArgs(args)
	if err != nil {
		return err
	}
	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}
	cr, err := circle.GetTreeContext(ctx, remote.Path, remote.RepoName, branch)
	if err != nil {
		return err
	}
	latestBuild := (*cr)[0]
	return circle.Rebuild(ctx, &latestBuild)
}

func main() {
	waitflags := flag.NewFlagSet("wait", flag.ExitOnError)
	waitRemote := waitflags.String("remote", "origin", "Git remote to use")
	waitflags.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: wait [refspec]

Wait for builds to complete, then print a descriptive output on success or
failure. By default, waits on the current branch, otherwise you can pass a
branch to wait for.

`)
		waitflags.PrintDefaults()
	}
	enableflags := flag.NewFlagSet("enable", flag.ExitOnError)
	enableflags.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", enableUsage)
		enableflags.PrintDefaults()
	}
	openflags := flag.NewFlagSet("open", flag.ExitOnError)
	downloadflags := flag.NewFlagSet("download-artifacts", flag.ExitOnError)
	downloadflags.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", downloadUsage)
		downloadflags.PrintDefaults()
	}
	rebuildflags := flag.NewFlagSet("rebuild", flag.ExitOnError)
	rebuildflags.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: rebuild

Rebuild a given test branch
`)
		rebuildflags.PrintDefaults()
	}

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
		return
	}
	subargs := args[1:]
	switch flag.Arg(0) {
	case "enable":
		enableflags.Parse(subargs)
		err := doEnable(enableflags)
		checkError(err)
	case "open":
		openflags.Parse(subargs)
		doOpen(openflags)
	case "rebuild":
		rebuildflags.Parse(subargs)
		err := doRebuild(rebuildflags)
		checkError(err)
	case "update":
		err := equinoxUpdate()
		checkError(err)
	case "version":
		fmt.Fprintf(os.Stderr, "circle version %s\n", circle.VERSION)
		os.Exit(1)
	case "wait":
		waitflags.Parse(subargs)
		args := waitflags.Args()
		branch, err := getBranchFromArgs(args)
		checkError(err)
		err = wait.Wait(branch, *waitRemote)
		checkError(err)
	case "download-artifacts":
		if len(args) == 1 {
			fmt.Fprintf(os.Stderr, "usage: download-artifacts <build-number>\n")
			os.Exit(1)
		}
		downloadflags.Parse(subargs)
		err := doDownload(downloadflags)
		checkError(err)
	default:
		usage()
	}
}
