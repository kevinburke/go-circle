package bigtext

import (
	"fmt"
	"os/exec"
)

const Version = "0.2"

// A Client is a vehicle for displaying information. Some properties may not
// be available in different notifiers
type Client struct {
	// Name is the name of the user or thing sending the command.
	Name string
	// LogoURL is the image to display alongside the notification.
	LogoURL string

	// ImageURL displays alongside to the text. With terminal-notifier, this is
	// to the right of the text.
	ImageURL string

	// OpenURL is a URL to open when you click on the notification.
	OpenURL string
}

var DefaultClient = Client{
	Name: "Terminal",
}

func (c *Client) Display(text string) error {
	if _, err := exec.LookPath("terminal-notifier"); err == nil {
		args := []string{"-title", c.Name, "-message", text}
		if c.LogoURL != "" {
			args = append(args, "-appIcon", c.LogoURL)
		}
		if c.ImageURL != "" {
			args = append(args, "-contentImage", c.ImageURL)
		}
		if c.OpenURL != "" {
			args = append(args, "-open", c.OpenURL)
		}
		_, err := exec.Command("terminal-notifier", args...).Output()
		return err
	} else {
		args := []string{
			"-e",
			fmt.Sprintf("tell application \"Quicksilver\" to show large type \"%s\"", text),
		}
		_, err := exec.Command("osascript", args...).Output()
		return err
	}
}

// Display text in large type. We try to use the terminal-notifier app if it's
// present, and try Quicksilver.app if terminal-notifier is not present. If
// neither application is present, an error is returned.
func Display(text string) error {
	return DefaultClient.Display(text)
}
