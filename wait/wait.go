package wait

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/Shyp/go-circle"
	"github.com/Shyp/go-git"
	"github.com/kevinburke/bigtext"
)

func roundDuration(d time.Duration, unit time.Duration) time.Duration {
	return ((d + unit/2) / unit) * unit
}

func round(f float64) int {
	if f < 0 {
		return int(f - 0.5)
	}
	return int(f + 0.5)
}

// getEffectiveCost returns the cost in cents to pay an average San
// Francisco-based engineer to wait for the amount of time specified by d.
func getEffectiveCost(d time.Duration) int {
	// https://www.glassdoor.com/Salaries/san-francisco-software-engineer-salary-SRCH_IL.0,13_IM759_KO14,31.htm
	yearlySalaryCents := float64(110554 * 100)
	// Estimate fully loaded costs add 40%.
	fullyLoadedSalary := yearlySalaryCents * 1.4

	workingDays := float64(49 * 5)
	hoursInWorkday := float64(8)
	salaryPerHour := fullyLoadedSalary * float64(d) / (workingDays * hoursInWorkday * float64(time.Hour))
	return round(salaryPerHour)
}

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

// getMinTipLength compares two git hashes and returns the length of the
// shortest
func getMinTipLength(remoteTip string, localTip string) int {
	var minTipLength int
	if len(remoteTip) <= len(localTip) {
		minTipLength = len(remoteTip)
	} else {
		minTipLength = len(localTip)
	}
	return minTipLength
}

func Wait(branch string) error {
	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}
	tip, err := git.Tip(branch)
	if err != nil {
		return err
	}
	fmt.Println("Waiting for latest build on", branch, "to complete")
	// Give CircleCI a little bit of time to start
	time.Sleep(1 * time.Second)
	for {
		cr, err := circle.GetTree(remote.Path, remote.RepoName, branch)
		if err != nil {
			if isHttpError(err) {
				fmt.Printf("Caught network error: %s. Continuing\n", err.Error())
				time.Sleep(2 * time.Second)
				continue
			}
			return err
		}
		if len(*cr) == 0 {
			return fmt.Errorf("No results, are you sure there are tests for %s/%s?\n",
				remote.Path, remote.RepoName)
		}
		latestBuild := (*cr)[0]
		maxTipLengthToCompare := getMinTipLength(latestBuild.VCSRevision, tip)
		if latestBuild.VCSRevision[:maxTipLengthToCompare] != tip[:maxTipLengthToCompare] {
			fmt.Printf("Latest build in Circle is %s, waiting for %s...\n",
				latestBuild.VCSRevision[:maxTipLengthToCompare], tip[:maxTipLengthToCompare])
			time.Sleep(5 * time.Second)
			continue
		}
		var duration time.Duration
		if latestBuild.QueuedAt.Valid {
			if latestBuild.StopTime.Valid {
				duration = latestBuild.StopTime.Time.Sub(latestBuild.QueuedAt.Time)
			} else {
				duration = time.Now().Sub(latestBuild.QueuedAt.Time)
			}
		} else if latestBuild.UsageQueuedAt.Valid {
			if latestBuild.StopTime.Valid {
				duration = latestBuild.StopTime.Time.Sub(latestBuild.UsageQueuedAt.Time)
			} else {
				duration = time.Now().Sub(latestBuild.UsageQueuedAt.Time)
			}
		}
		duration = roundDuration(duration, time.Second)
		if latestBuild.Passed() {
			fmt.Printf("Build on %s succeeded!\n\n", branch)
			build, err := circle.GetBuild(remote.Path, remote.RepoName, latestBuild.BuildNum)
			if err == nil {
				fmt.Print(build.Statistics())
			} else {
				fmt.Printf("error getting build: %v\n", err)
			}
			fmt.Printf("\nTests on %s took %s. Quitting.\n", branch, duration.String())
			c := bigtext.Client{
				Name:    fmt.Sprintf("%s (go-circle)", remote.RepoName),
				OpenURL: latestBuild.BuildURL,
			}
			c.Display(branch + " build complete!")
			break
		} else if latestBuild.Failed() {
			build, err := circle.GetBuild(remote.Path, remote.RepoName, latestBuild.BuildNum)
			if err == nil {
				fmt.Print(build.Statistics())
				texts, textsErr := build.FailureTexts(context.Background())
				if textsErr != nil {
					fmt.Printf("error getting build failures: %v\n", textsErr)
				}
				fmt.Printf("\nOutput from failed builds:\n\n")
				for _, text := range texts {
					fmt.Println(text)
				}
			} else {
				fmt.Printf("error getting build: %v\n", err)
			}
			fmt.Printf("\nURL: %s\n", latestBuild.BuildURL)
			err = fmt.Errorf("Build on %s failed!\n\n", branch)
			c := bigtext.Client{
				Name:    fmt.Sprintf("%s (go-circle)", remote.RepoName),
				OpenURL: latestBuild.BuildURL,
			}
			c.Display("build failed")
			return err
		} else {
			if latestBuild.Status == "running" {
				fmt.Printf("Running (%s elapsed)\n", duration.String())
			} else if latestBuild.NotRunning() {
				cost := getEffectiveCost(duration)
				centsPortion := cost % 100
				dollarPortion := cost / 100
				costStr := fmt.Sprintf("$%d.%.2d", dollarPortion, centsPortion)
				fmt.Printf("Status is %s (queued for %s, cost %s), trying again\n",
					latestBuild.Status, duration.String(), costStr)
			} else {
				fmt.Printf("Status is %s, trying again\n", latestBuild.Status)
			}
			// Sleep less and less as we approach the duration of the previous
			// successful build
			buildDuration := time.Duration(latestBuild.Previous.BuildDurationMs) * time.Millisecond
			if latestBuild.Previous.Status == "success" || latestBuild.Previous.Status == "fixed" {
				if duration < time.Minute {
					// First minute, errors are slightly more likely.
					time.Sleep(5 * time.Second)
				} else {
					timeRemaining := buildDuration - duration
					if timeRemaining > 5*time.Minute {
						time.Sleep(30 * time.Second)
					} else if timeRemaining > 3*time.Minute {
						time.Sleep(20 * time.Second)
					} else if timeRemaining > time.Minute {
						time.Sleep(15 * time.Second)
					} else if timeRemaining > 30*time.Second {
						time.Sleep(10 * time.Second)
					} else if timeRemaining > 10*time.Second {
						time.Sleep(5 * time.Second)
					} else {
						time.Sleep(3 * time.Second)
					}
				}
			} else {
				if float32(duration) < (2.5 * float32(time.Minute)) {
					time.Sleep(10 * time.Second)
				} else {
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
	return nil
}
