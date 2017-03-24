package circle

import (
	"fmt"
	"testing"
)

func TestBuild(t *testing.T) {
	t.Skip()
	build, err := GetBuild("Shyp", "shyp_api", 15523)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(build.Statistics())
}
