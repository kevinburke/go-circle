package circle

import (
	"fmt"
	"testing"
)

func TestBuild(t *testing.T) {
	t.Skip()
	build, err := GetBuild("github.com", "kevinburke", "go-circle", 15523)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(build.Statistics())
}
