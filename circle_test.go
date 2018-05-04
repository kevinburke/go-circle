package circle

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestBuild(t *testing.T) {
	t.Skip()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	build, err := GetBuild(ctx, "github.com", "kevinburke", "go-circle", 15523)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(build.Statistics(false))
}
