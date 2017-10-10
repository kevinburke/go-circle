package build

import (
	"fmt"

	"github.com/Shyp/go-git"
	"github.com/kevinburke/go-circle"
)

// GetBuilds gets the status of the 5 most recent Circle builds for a branch
func GetBuilds(branch string) error {
	// This throws if the branch doesn't exist
	if _, err := git.Tip(branch); err != nil {
		return err
	}

	fmt.Printf("\nFetching recent builds for %s starting with most recent commit\n\n", branch)

	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}

	cr, err := circle.GetTree(remote.Path, remote.RepoName, branch)
	if err != nil {
		return err
	}

	buildCount := 0

	if len(*cr) < 5 {
		buildCount = len(*cr)
	} else {
		buildCount = 5 // Limited to 5 most recent builds.
	}

	for i := 0; i < buildCount; i++ {
		build := (*cr)[i]
		ghUrl, url, status := build.CompareURL, build.BuildURL, build.Status

		// Based on the status of the build, change the color of status print out
		if build.Passed() {
			status = fmt.Sprintf("\033[38;05;119m%-8s\033[0m", status)
		} else if build.NotRunning() {
			status = fmt.Sprintf("\033[38;05;20m%-8s\033[0m", status)
		} else if build.Failed() {
			status = fmt.Sprintf("\033[38;05;160m%-8s\033[0m", status)
		} else if build.Running() {
			status = fmt.Sprintf("\033[38;05;80m%-8s\033[0m", status)
		} else {
			status = fmt.Sprintf("\033[38;05;0m%-8s\033[0m", status)
		}

		fmt.Println(url, status, ghUrl)

	}

	fmt.Println("\nMost recent build statuses fetched!")

	return nil
}

// CancelBuild cancels a build (as specified by the build number)
func CancelBuild(org string, project string, buildNum int) string {
	fmt.Printf("\nCanceling build: %d for %s\n\n", buildNum, project)
	if _, err := circle.CancelBuild(org, project, buildNum); err != nil {
		return ""
	}

	return ""
}
