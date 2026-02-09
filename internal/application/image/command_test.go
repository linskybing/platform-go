package image

import (
	"testing"
)

func TestCraneCopyArgs(t *testing.T) {
	src := "docker.io/library/nginx:latest"
	dst := "harbor-prefix/library/nginx:latest"
	args := CraneCopyArgs(src, dst)
	if len(args) != 5 {
		t.Fatalf("unexpected args length: %d", len(args))
	}
	if args[0] != "crane" || args[1] != "copy" || args[2] != src || args[3] != dst || args[4] != "--insecure" {
		t.Fatalf("unexpected crane args: %v", args)
	}
}
