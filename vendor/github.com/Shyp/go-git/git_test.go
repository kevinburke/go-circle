package git

import (
	"reflect"
	"testing"
)

func TestCurrentBranch(t *testing.T) {
	result, err := CurrentBranch()
	// TODO find a way to test this that does not rely on the current state of
	// the git repository.
	_ = err
	_ = result
}

func TestRoot(t *testing.T) {
	result, err := Root()
	// TODO figure out a way to test this as well - it depends on your current
	// working directory, and you can run "go test" from anywhere on the
	// filesystem.
	_ = err
	_ = result
}

var remoteTests = []struct {
	remote   string
	expected RemoteURL
}{
	{
		"git@github.com:Shyp/shyp_api.git", RemoteURL{
			Host:     "github.com",
			Port:     22,
			Path:     "Shyp",
			RepoName: "shyp_api",
			Format:   SSHFormat,
			URL:      "git@github.com:Shyp/shyp_api.git",
			SSHUser:  "git",
		},
	}, {
		"git@github.com:Shyp/shyp_api.git/", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     22,
			RepoName: "shyp_api",
			Format:   SSHFormat,
			URL:      "git@github.com:Shyp/shyp_api.git/",
			SSHUser:  "git",
		},
	}, {
		"git@github.com:path/to/Shyp/shyp_api.git/", RemoteURL{
			Path:     "path/to/Shyp",
			Host:     "github.com",
			Port:     22,
			RepoName: "shyp_api",
			Format:   SSHFormat,
			URL:      "git@github.com:path/to/Shyp/shyp_api.git/",
			SSHUser:  "git",
		},
	}, {
		"https://github.com/Shyp/shyp_api.git", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     443,
			RepoName: "shyp_api",
			Format:   HTTPSFormat,
			URL:      "https://github.com/Shyp/shyp_api.git",
			SSHUser:  "",
		},
	}, {
		"https://github.com/Shyp/shyp_api.git/", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     443,
			RepoName: "shyp_api",
			Format:   HTTPSFormat,
			URL:      "https://github.com/Shyp/shyp_api.git/",
			SSHUser:  "",
		},
	}, {
		"https://github.com:11443/Shyp/shyp_api.git", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     11443,
			RepoName: "shyp_api",
			Format:   HTTPSFormat,
			URL:      "https://github.com:11443/Shyp/shyp_api.git",
			SSHUser:  "",
		},
	}, {
		"https://github.com/Shyp/shyp_api", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     443,
			RepoName: "shyp_api",
			Format:   HTTPSFormat,
			URL:      "https://github.com/Shyp/shyp_api",
			SSHUser:  "",
		},
	}, {
		"https://github.com/Shyp/repo.name.with.periods.git", RemoteURL{
			Path:     "Shyp",
			Host:     "github.com",
			Port:     443,
			RepoName: "repo.name.with.periods",
			Format:   HTTPSFormat,
			URL:      "https://github.com/Shyp/repo.name.with.periods.git",
			SSHUser:  "",
		},
	},
}

func TestParseRemoteURL(t *testing.T) {
	for _, tt := range remoteTests {
		remote, err := ParseRemoteURL(tt.remote)
		if err != nil {
			t.Fatal(err)
		}
		if remote == nil {
			t.Fatalf("expected ParseRemoteURL(%s) to be %v, was nil", tt.remote, tt.expected)
		}
		if !reflect.DeepEqual(*remote, tt.expected) {
			t.Errorf("expected ParseRemoteURL(%s) to be %#v, was %#v", tt.remote, tt.expected, remote)
		}
	}
}
