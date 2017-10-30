package circle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	types "github.com/kevinburke/go-types"
	"github.com/kevinburke/rest"
	"golang.org/x/sync/errgroup"
)

var client http.Client

// TODO switch all clients to use this
var v11client *rest.Client

func init() {
	client = http.Client{
		Timeout: 10 * time.Second,
	}
	v11client = rest.NewClient("", "", v11BaseUri)
	// use Context to set a timeout on this client
	v11client.Client.Timeout = 0
}

const VERSION = "0.25"
const baseUri = "https://circleci.com/api/v1/project"
const v11BaseUri = "https://circleci.com/api/v1.1/project"

type TreeBuild struct {
	BuildNum   int    `json:"build_num"`
	BuildURL   string `json:"build_url"`
	CompareURL string `json:"compare"`
	// Tree builds have a `previous_successful_build` field but as far as I can
	// tell it is always null. Instead this field is set
	Previous      PreviousBuild  `json:"previous"`
	QueuedAt      types.NullTime `json:"queued_at"`
	RepoName      string         `json:"reponame"`
	Status        string         `json:"status"`
	StartTime     types.NullTime `json:"start_time"`
	StopTime      types.NullTime `json:"stop_time"`
	UsageQueuedAt types.NullTime `json:"usage_queued_at"`
	Username      string         `json:"username"`
	VCSRevision   string         `json:"vcs_revision"`
	VCSType       string         `json:"vcs_type"`
}

func (tb TreeBuild) Passed() bool {
	return tb.Status == "success" || tb.Status == "fixed"
}

func (tb TreeBuild) NotRunning() bool {
	return tb.Status == "not_running" || tb.Status == "scheduled" || tb.Status == "queued"
}

func (tb TreeBuild) Running() bool {
	return tb.Status == "running"
}

func (tb TreeBuild) Failed() bool {
	return tb.Status == "failed" || tb.Status == "timedout" || tb.Status == "no_tests" || tb.Status == "infrastructure_fail"
}

type CircleArtifact struct {
	Path       string `json:"path"`
	PrettyPath string `json:"pretty_path"`
	NodeIndex  uint8  `json:"node_index"`
	Url        string `json:"url"`
}

type CircleBuild struct {
	BuildNum                uint32         `json:"build_num"`
	Parallel                uint8          `json:"parallel"`
	Platform                string         `json:"platform"`
	PreviousSuccessfulBuild PreviousBuild  `json:"previous_successful_build"`
	QueuedAt                types.NullTime `json:"queued_at"`
	RepoName                string         `json:"reponame"` // "go"
	Steps                   []Step         `json:"steps"`
	VCSType                 string         `json:"vcs_type"` // "github", "bitbucket"
	UsageQueuedAt           types.NullTime `json:"usage_queued_at"`
	Username                string         `json:"username"` // "golang"
}

// Failures returns an array of (buildStep, containerID) integers identifying
// the IDs of container/build step pairs that failed.
func (cb CircleBuild) Failures() [][2]int {
	failures := make([][2]int, 0)
	for i, step := range cb.Steps {
		for j, action := range step.Actions {
			if action.Failed() {
				if cb.Platform == "2.0" {
					failures = append(failures, [...]int{action.Step, j})
				} else {
					failures = append(failures, [...]int{i, j})
				}
			}
		}
	}
	return failures
}

type CircleOutput struct {
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
	Type    string    `json:"type"`
}

type CircleOutputs []*CircleOutput

func (cb CircleBuild) FailureTexts(ctx context.Context) ([]string, error) {
	group, errctx := errgroup.WithContext(ctx)
	// todo this is not great design
	token, err := getToken(cb.Username)
	if err != nil {
		return nil, err
	}
	// V2: https://circleci.com/api/v1.1/project/github/kevinburke/go-circle/8/output/102/0?allocation-id=59dc1a06c9e77c0001793e56-0-build%2F346CFC34&truncate=400000
	failures := cb.Failures()
	results := make([]string, len(failures))
	for i, failure := range failures {
		failure := failure
		i := i
		// https://circleci.com/api/v1.1/project/github/segmentio/terracode-bootstrap/11/output/9/0
		group.Go(func() error {
			uri := fmt.Sprintf("/%s/%s/%s/%d/output/%d/%d?circle-token=%s", cb.VCSType, cb.Username, cb.RepoName, cb.BuildNum, failure[0], failure[1], token)
			req, err := v11client.NewRequest("GET", uri, nil)
			if err != nil {
				return err
			}
			req = req.WithContext(errctx)
			var outputs []*CircleOutput
			if err := v11client.Do(req, &outputs); err != nil {
				return err
			}
			var message string
			for i := range outputs {
				message = message + outputs[i].Message + "\n"
			}
			results[i] = message
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		return nil, err
	}
	return results, nil
}

type PreviousBuild struct {
	BuildNum int `json:"build_num"`
	// would be neat to make this a time.Duration, easier to use the passed in
	// value.
	Status string `json:"status"`

	BuildDurationMs int `json:"build_time_millis"`
}

type Step struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

type Action struct {
	Name         string         `json:"name"`
	AllocationID string         `json:"allocation_id"`
	Index        uint16         `json:"index"`
	OutputURL    URL            `json:"output_url"`
	Runtime      CircleDuration `json:"run_time_millis"`
	Status       string         `json:"status"`
	Step         int            `json:"step"`

	// Failed is a boolean, but we defined it already as a function on the
	// Action so for compat we should pick a different name when we destructure
	// it.
	HasFailed bool `json:"failed"`
}

func (a Action) Failed() bool {
	return a.HasFailed
}

func getTreeUri(org string, project string, branch string, token string) string {
	return fmt.Sprintf("/%s/%s/tree/%s?circle-token=%s", org, project, branch, token)
}

func getBuildUri(org string, project string, build int, token string) string {
	return fmt.Sprintf("/%s/%s/%d?circle-token=%s", org, project, build, token)
}

func getCancelUri(org string, project string, build int, token string) string {
	return fmt.Sprintf("%s/%s/%s/%d/cancel?circle-token=%s", baseUri, org, project, build, token)
}

func getArtifactsUri(org string, project string, build int, token string) string {
	return fmt.Sprintf("%s/%s/%s/%d/artifacts?circle-token=%s", baseUri, org, project, build, token)
}

type CircleTreeResponse []TreeBuild

func makeRequest(method, uri string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("circle-command-line-client/%s", VERSION))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		// TODO handle build not found, you get back {"message": "Build not found"}
		return nil, fmt.Errorf("Request failed with status [%d]", resp.StatusCode)
	}

	return resp.Body, nil
}

type FollowResponse struct {
	Following bool `json:"following"`
	// TODO...
}

func Enable(ctx context.Context, host string, org string, repoName string) error {
	token, err := getToken(org)
	if err != nil {
		return err
	}
	var vcs string
	switch {
	case strings.Contains(host, "github.com"):
		vcs = "github"
	case strings.Contains(host, "bitbucket.org"):
		vcs = "bitbucket"
	default:
		return fmt.Errorf("can't enable unknown host %s", host)
	}
	uri := fmt.Sprintf("/%s/%s/%s/follow?circle-token=%s", vcs, org, repoName, token)
	req, err := v11client.NewRequest("POST", uri, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	fr := new(FollowResponse)
	if err := v11client.Do(req, fr); err != nil {
		return err
	}
	if fr.Following == false {
		return errors.New("not following the project")
	}
	return nil
}

func Rebuild(ctx context.Context, tb *TreeBuild) error {
	token, err := getToken(tb.Username)
	if err != nil {
		return err
	}
	// https://circleci.com/gh/segmentio/db-service/1488
	// url we have is https://circleci.com/api/v1.1/project/github/segmentio/db-service/1486/retry
	uri := fmt.Sprintf("/%s/%s/%s/%d/retry?circle-token=%s", tb.VCSType, tb.Username, tb.RepoName, tb.BuildNum, token)
	req, err := v11client.NewRequest("POST", uri, strings.NewReader("null"))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	return v11client.Do(req, nil)
}

func GetTree(org string, project string, branch string) (*CircleTreeResponse, error) {
	return GetTreeContext(context.Background(), org, project, branch)
}

func GetTreeContext(ctx context.Context, org, project, branch string) (*CircleTreeResponse, error) {
	token, err := getToken(org)
	if err != nil {
		return nil, err
	}
	uri := getTreeUri(org, project, branch, token)
	client := rest.NewClient("", "", baseUri)
	req, err := client.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	cr := new(CircleTreeResponse)
	if err := client.Do(req, cr); err != nil {
		return nil, err
	}
	return cr, nil
}

func GetBuild(org string, project string, buildNum int) (*CircleBuild, error) {
	token, err := getToken(org)
	if err != nil {
		return nil, err
	}
	uri := getBuildUri(org, project, buildNum, token)
	client := rest.NewClient("", "", baseUri)
	req, err := client.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	cb := new(CircleBuild)
	if err := client.Do(req, cb); err != nil {
		return nil, err
	}
	return cb, nil
}

func GetArtifactsForBuild(org string, project string, buildNum int) ([]*CircleArtifact, error) {
	token, err := getToken(org)
	if err != nil {
		return []*CircleArtifact{}, err
	}
	uri := getArtifactsUri(org, project, buildNum, token)
	body, err := makeRequest("GET", uri)
	if err != nil {
		return []*CircleArtifact{}, err
	}
	defer body.Close()
	var r io.Reader
	if os.Getenv("CIRCLE_DEBUG") == "true" {
		r = io.TeeReader(body, os.Stdout)
	} else {
		r = body
	}
	var arts []*CircleArtifact
	decodeErr := json.NewDecoder(r).Decode(&arts)
	return arts, decodeErr
}

func DownloadArtifact(artifact *CircleArtifact, directory string, org string) error {
	token, err := getToken(org)
	if err != nil {
		return err
	}
	fname := fmt.Sprintf("%d.%s", artifact.NodeIndex, path.Base(artifact.Url))
	fmt.Fprintf(os.Stderr, "Downloading artifact to %s\n", fname)
	f, err := os.Create(filepath.Join(directory, fname))
	if err != nil {
		return err
	}
	defer f.Close()
	url := fmt.Sprintf("%s?circle-token=%s", artifact.Url, token)
	body, err := makeRequest("GET", url)
	if err != nil {
		return err
	}
	defer body.Close()
	_, copyErr := io.Copy(f, body)
	return copyErr
}

func CancelBuild(org string, project string, buildNum int) (*CircleBuild, error) {
	token, err := getToken(org)
	if err != nil {
		return nil, err
	}
	uri := getCancelUri(org, project, buildNum, token)
	body, err := makeRequest("POST", uri)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var r io.Reader
	if os.Getenv("CIRCLE_DEBUG") == "true" {
		r = io.TeeReader(body, os.Stdout)
	} else {
		r = body
	}
	d := json.NewDecoder(r)
	var cb CircleBuild
	err = d.Decode(&cb)
	return &cb, err
}
