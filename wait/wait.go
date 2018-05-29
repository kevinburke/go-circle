package wait

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kevinburke/bigtext"
	"github.com/kevinburke/go-circle"
	"github.com/kevinburke/go-git"
	"github.com/kevinburke/remoteci"
	"github.com/pkg/browser"
)

// isHttpError checks if the given error is a request timeout or a network
// failure - in those cases we want to just retry the request.
func isHttpError(err error) bool {
	if err == nil {
		return false
	}
	// some net.OpError's are wrapped in a url.Error
	if uerr, ok := err.(*url.Error); ok {
		err = uerr.Err
	}
	switch err := err.(type) {
	default:
		return false
	case *net.OpError:
		return err.Op == "dial" && err.Net == "tcp"
	case *net.DNSError:
		return true
	// Catchall, this needs to go last.
	case net.Error:
		return err.Timeout() || err.Temporary()
	}
}

var debugCmd = os.Getenv("DEBUG_CMD") == "true"

func runCmd(ctx context.Context, stdout, stderr io.Writer, cmdName string, args ...string) error {
	cmd := exec.CommandContext(ctx, cmdName, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if debugCmd {
		fmt.Fprintln(os.Stderr, strings.Join(cmd.Args, " "))
	}
	return cmd.Run()
}

// getShorterString compares two git hashes and returns the length of the
// shortest
func getShorterString(a string, b string) int {
	if len(a) <= len(b) {
		return len(a)
	}
	return len(b)
}

func rebase(ctx context.Context, branch, remoteStr, rebaseAgainst string, c *bigtext.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	buf := new(bytes.Buffer)
	remoteRef := remoteStr + "/" + rebaseAgainst
	if err := runCmd(ctx, buf, buf, "git", "rebase", remoteRef, branch); err != nil {
		abortBuf := new(bytes.Buffer)
		abortErr := runCmd(ctx, abortBuf, abortBuf, "git", "rebase", "--abort")
		fmt.Printf(`Remote branch %s changed, and automatic rebase failed.

Rebase output was:

%s
`, remoteRef, buf.String())
		if abortErr != nil {
			fmt.Printf(`Attempted to abort, but abort failed. git rebase --abort output was:

%s
`, abortBuf)
		} else {
			fmt.Printf("Rebase was aborted. Quitting.\n")
		}
		c.Display("rebase failed")
		return err
	}
	buf.Reset()
	if err := forcePush(ctx, branch, remoteStr, buf); err != nil {
		fmt.Printf(`Remote branch %s changed, we performed a local rebase but push failed.

Push output was:

%s
`, remoteRef, buf.String())
		c.Display("push after rebase failed")
		return err
	}
	return nil
}

func forcePush(ctx context.Context, branch, remoteStr string, w io.Writer) error {
	return runCmd(ctx, w, w, "git", "push", "--force-with-lease", remoteStr, branch+":"+branch)
}

func newRemoteCommit(ctx context.Context, branch, remoteStr, remoteBranch string) (string, error) {
	remoteRef := remoteStr + "/" + remoteBranch
	buf := new(bytes.Buffer)
	if err := runCmd(ctx, buf, buf, "git", "fetch", remoteStr); err != nil {
		return "", err
	}
	buf.Reset()
	commitBuf := new(strings.Builder)
	if err := runCmd(ctx, commitBuf, buf, "git", "show-ref", "--hash", remoteRef); err != nil {
		return "", err
	}
	buf.Reset()
	mergeBaseBuf := new(strings.Builder)
	if err := runCmd(ctx, mergeBaseBuf, buf, "git", "merge-base", remoteRef, branch); err != nil {
		return "", err
	}
	commit := strings.TrimSpace(commitBuf.String())
	mergeBase := strings.TrimSpace(mergeBaseBuf.String())
	if commit != mergeBase {
		return commit, nil
	}
	return "", nil
}

var debugHTTPTraffic = os.Getenv("DEBUG_HTTP_TRAFFIC") == "true"

func clear(w io.Writer, lines int) {
	if !debugHTTPTraffic {
		io.WriteString(w, strings.Repeat("\033[2K\r\033[1A", lines))
	}
}

func draw(w io.Writer, build *circle.CircleBuild, prevLinesDrawn int) int {
	stats := build.Statistics(true)
	clear(w, prevLinesDrawn)
	io.WriteString(w, stats+"\n\033[?25l")
	return strings.Count(stats, "\n") + 1
}

func isCtxCanceled(err error) bool {
	if err == nil {
		return false
	}
	if err == context.Canceled {
		return true
	}
	uerr, ok := err.(*url.Error)
	if ok && uerr.Err == context.Canceled {
		return true
	}
	return false
}

func wait(ctx context.Context, branch, remoteStr string, rebaseAgainst string) error {
	tty := remoteci.IsATTY(os.Stdout)
	if tty {
		defer func() {
			fmt.Printf("\033[?25h")
		}()
	}
	remote, err := git.GetRemoteURL(remoteStr)
	if err != nil {
		return err
	}
	tip, err := git.Tip(branch)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	waitCtx, waitCancel := context.WithCancel(ctx)
	defer waitCancel()
	var newCommit string
	var newCommitMu sync.Mutex
	if rebaseAgainst != "" {
		ticker := time.NewTicker(10 * time.Second)
		wg.Add(1)
		defer ticker.Stop()
		go func() {
			runOnce := false
			for {
				ctx, cancel := context.WithTimeout(waitCtx, 1*time.Minute)
				commit, err := newRemoteCommit(ctx, branch, remoteStr, rebaseAgainst)
				cancel()
				if err != nil || commit == "" {
					// nothing to do
				} else {
					newCommitMu.Lock()
					newCommit = commit
					newCommitMu.Unlock()
				}
				if !runOnce {
					wg.Done()
					runOnce = true
				}
				select {
				case <-waitCtx.Done():
					return
				case <-ticker.C:
				}
			}
		}()
	}
	checkRebase := func(c *bigtext.Client) error {
		if rebaseAgainst != "" {
			newCommitMu.Lock()
			commit := newCommit
			newCommitMu.Unlock()
			if commit != "" {
				// commit does not match merge base
				fmt.Printf("Remote branch %s/%s has changed, rebasing %s on top of it\n", remoteStr, rebaseAgainst, branch)
				if err := rebase(waitCtx, branch, remoteStr, rebaseAgainst, c); err != nil {
					return err
				}
				newCommitMu.Lock()
				newCommit = ""
				newCommitMu.Unlock()
				return errChangedRemote
			}
		}
		return nil
	}
	fmt.Println("Waiting for latest build on", branch, "to complete")
	// Give CircleCI a little bit of time to start
	select {
	case <-waitCtx.Done():
		return nil
	case <-time.After(1 * time.Second):
	}
	var lastPrintedAt time.Time
	linesDrawn := 0
	hasOpenedFailedBuild := false
	for {
		cr, err := circle.GetTreeContext(waitCtx, remote.Host, remote.Path, remote.RepoName, branch)
		if err != nil {
			if isCtxCanceled(err) {
				return nil
			}
			if isHttpError(err) {
				fmt.Printf("Caught network error: %s. Continuing\n", err.Error())
				lastPrintedAt = time.Now()
				select {
				case <-waitCtx.Done():
					return nil
				case <-time.After(2 * time.Second):
				}
				continue
			}
			return err
		}
		if len(*cr) == 0 {
			return fmt.Errorf("No results, are you sure there are tests for %s/%s?\n",
				remote.Path, remote.RepoName)
		}
		var detailedBuild *circle.CircleBuild
		latestBuild := (*cr)[0]
		c := bigtext.Client{
			Name:    fmt.Sprintf("%s (go-circle)", remote.RepoName),
			OpenURL: latestBuild.BuildURL,
		}
		maxTipLengthToCompare := getShorterString(latestBuild.VCSRevision, tip)
		tip = tip[:maxTipLengthToCompare]
		shortVCSRev := latestBuild.VCSRevision[:maxTipLengthToCompare]
		if shortVCSRev != tip {
			if rebaseAgainst != "" {
				buf := new(bytes.Buffer)
				if err := forcePush(waitCtx, branch, remoteStr, buf); err != nil {
					fmt.Printf(`CircleCI built commit %s does not match local commit %s.

We attempted a force push to %s/%s to trigger a build, but it failed.
Push output was:

%s
`, shortVCSRev, tip, remoteStr, branch, buf.String())
					c.Display("force push failed")
					return err
				}
				fmt.Printf("Force pushed local commit %s to %s/%s to trigger new build...\n", tip, remoteStr, branch)
				// shortest CircleCI build I've ever seen is 20 seconds, so we
				// have some time to wait before a complete build.
				select {
				case <-waitCtx.Done():
					return nil
				case <-time.After(7 * time.Second):
				}
			} else {
				fmt.Printf("Latest build in Circle is %s, waiting for %s...\n",
					shortVCSRev, tip)
				lastPrintedAt = time.Now()
				select {
				case <-waitCtx.Done():
					return nil
				case <-time.After(5 * time.Second):
				}
			}
			continue
		}
		duration := latestBuild.Elapsed().Round(time.Second)
		if latestBuild.Passed() {
			wg.Wait()
			if err := checkRebase(&c); err != nil {
				return err
			}
			var err error
			detailedBuild, err = circle.GetBuild(waitCtx, remote.Host, remote.Path, remote.RepoName, latestBuild.BuildNum)
			switch {
			case err != nil:
				fmt.Printf("error getting build statistics: %v\n", err)
			case tty:
				// need one last draw with the final timings
				draw(os.Stdout, detailedBuild, linesDrawn)
				clear(os.Stdout, 1)
			default:
				fmt.Print(detailedBuild.Statistics(false))
			}
			fmt.Printf(`Build on %s succeeded!

Tests on %s took %s. Quitting.
`, branch, branch, duration.String())
			c.Display(branch + " build complete!")
			break
		}
		if latestBuild.Failed() {
			wg.Wait()
			if err := checkRebase(&c); err != nil {
				return err
			}
			var err error
			detailedBuild, err = circle.GetBuild(waitCtx, remote.Host, remote.Path, remote.RepoName, latestBuild.BuildNum)
			switch {
			case err != nil:
				fmt.Printf("error getting build stats: %v\n", err)
				return err
			case tty:
				draw(os.Stdout, detailedBuild, linesDrawn)
				clear(os.Stdout, 1)
			default:
				fmt.Print(detailedBuild.Statistics(false))
			}
			failureCtx, cancel := context.WithTimeout(waitCtx, 20*time.Second)
			texts, textsErr := detailedBuild.FailureTexts(failureCtx)
			if textsErr != nil {
				fmt.Printf("error getting build failures: %v\n", textsErr)
			}
			cancel()
			fmt.Printf("\nOutput from failed builds:\n\n")
			for i := range texts {
				fmt.Println(texts[i])
			}
			fmt.Printf("\nURL: %s\n", latestBuild.BuildURL)
			err = fmt.Errorf("Build on %s failed!\n\n", branch)
			c.Display("build failed")
			return err
		}
		if latestBuild.Status == "running" {
			build, err := circle.GetBuild(waitCtx, remote.Host, remote.Path, remote.RepoName, latestBuild.BuildNum)
			if err != nil {
				if isCtxCanceled(err) {
					return nil
				}
				// draw one extra line
				fmt.Printf("Caught network error: %s. Continuing\n", err.Error())
				linesDrawn++
			} else {
				if tty {
					linesDrawn = draw(os.Stdout, build, linesDrawn)
				} else {
					// use the elapsed duration for predicting how long the build will
					// take to complete, but print the duration - we should show users
					// the time since their build was pushed, not when Circle decided to
					// start running it.
					fmt.Printf("Build %d running (%s elapsed)\n", latestBuild.BuildNum, duration.String())
					linesDrawn++
				}
				if !hasOpenedFailedBuild {
					wg.Wait()
					if err := checkRebase(&c); err != nil {
						return err
					}
					// todo logic like this also exists in circle/main.go
					for _, step := range build.Steps {
						for _, action := range step.Actions {
							if action.Failed() {
								u := latestBuild.BuildURL + "#tests/containers/" + strconv.FormatUint(uint64(action.Index), 10)
								if err := browser.OpenURL(u); err == nil {
									hasOpenedFailedBuild = true
								}
							}
						}
					}
				}
			}
		} else if latestBuild.NotRunning() {
			wg.Wait()
			if err := checkRebase(&c); err != nil {
				return err
			}
			cost := remoteci.GetEffectiveCost(duration)
			centsPortion := cost % 100
			dollarPortion := cost / 100
			costStr := fmt.Sprintf("$%d.%.2d", dollarPortion, centsPortion)
			if lastPrintedAt.Add(12 * time.Second).Before(time.Now()) {
				fmt.Printf("Status is %s (queued for %s, cost %s), trying again\n",
					latestBuild.Status, duration.String(), costStr)
				lastPrintedAt = time.Now()
			}
		} else {
			fmt.Printf("Status is %s, trying again\n", latestBuild.Status)
			lastPrintedAt = time.Now()
		}
		if err := checkRebase(&c); err != nil {
			return err
		}
		sleepCh := time.After(3 * time.Second)
		stillSleeping := true
		for stillSleeping {
			select {
			case <-waitCtx.Done():
				return nil
			case <-sleepCh:
				stillSleeping = false
			case <-time.After(200 * time.Millisecond):
				if latestBuild.Status == "running" {
					clear(os.Stdout, 2)
					fmt.Fprintf(os.Stdout, "Build %d running... %s elapsed\n\n", latestBuild.BuildNum, latestBuild.Elapsed().Round(time.Second))
				}
			}
		}
	}
	return nil
}

var errChangedRemote = errors.New("remote branch changed")

// Wait waits for a build on the local branch to finish in CircleCI. If
// rebaseAgainst is not empty, Wait will periodically fetch that branch from the
// remote and rebase against it if it changes.
func Wait(ctx context.Context, branch, remoteStr string, rebaseAgainst string) error {
	for {
		err := wait(ctx, branch, remoteStr, rebaseAgainst)
		if err == errChangedRemote {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(7 * time.Second):
			}
			continue
		}
		return err
	}
}
