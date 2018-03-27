package circle

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type CircleConfig struct {
	Organizations map[string]organization
}

type organization struct {
	Token string
}

// getCaseInsensitiveOrg finds the key in the list of orgs. This is a case
// insensitive match, so if key is "ShyP" and orgs has a key named "sHYp",
// that will count as a match.
func getCaseInsensitiveOrg(key string, orgs map[string]organization) (organization, error) {
	for k, _ := range orgs {
		lower := strings.ToLower(k)
		if _, ok := orgs[lower]; !ok {
			orgs[lower] = orgs[k]
			delete(orgs, k)
		}
	}
	lowerKey := strings.ToLower(key)
	if o, ok := orgs[lowerKey]; ok {
		return o, nil
	} else {
		return organization{}, fmt.Errorf(`Couldn't find organization %s in the config.

Go to https://circleci.com/account/api if you need to create a token.
`, key)
	}
}

func getToken(orgName string) (string, error) {
	var filename string
	var f io.ReadCloser
	var err error
	checkedLocations := make([]string, 1)
	if cfg, ok := os.LookupEnv("XDG_CONFIG_HOME"); ok {
		filename = filepath.Join(cfg, "circleci")
		f, err = os.Open(filename)
		checkedLocations[0] = filename
	} else {
		var homeDir string
		user, userErr := user.Current()
		if userErr == nil {
			homeDir = user.HomeDir
		} else {
			homeDir = os.Getenv("HOME")
		}
		filename = filepath.Join(homeDir, "cfg", "circleci")
		f, err = os.Open(filename)
		checkedLocations[0] = filename
		if err != nil { //fallback
			rcFilename := filepath.Join(homeDir, ".circlerc")
			f, err = os.Open(rcFilename)
			checkedLocations = append(checkedLocations, rcFilename)
		}
	}
	if err != nil {
		err = fmt.Errorf(`Couldn't find a config file in %s.

Add a configuration file with your CircleCI token, like this:

[organizations]

    [organizations.Shyp]
    token = "aabbccddeeff00"

Go to https://circleci.com/account/api if you need to create a token.
`, strings.Join(checkedLocations, " or "))
		return "", err
	}
	defer f.Close()
	var c CircleConfig
	_, err = toml.DecodeReader(bufio.NewReader(f), &c)
	if err != nil {
		return "", err
	}
	org, err := getCaseInsensitiveOrg(orgName, c.Organizations)
	if err != nil {
		return "", err
	}
	return org.Token, nil
}
