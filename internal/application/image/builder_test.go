package image

import (
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
)

func TestBuildPullJob_NormalizeAndCommand(t *testing.T) {
	name := "nginx"
	tag := "1.2.3"
	job, full, harbor := BuildPullJob(name, tag)
	if job == nil {
		t.Fatal("expected job, got nil")
	}

	if len(job.Spec.Template.Spec.InitContainers) == 0 {
		t.Fatalf("no init containers in job: %+v", job)
	}
	initImg := job.Spec.Template.Spec.InitContainers[0].Image
	if !strings.HasPrefix(initImg, "docker.io/library/") {
		t.Fatalf("init image not normalized: %s", initImg)
	}

	if len(job.Spec.Template.Spec.Containers) == 0 {
		t.Fatalf("no containers in job")
	}
	cmd := job.Spec.Template.Spec.Containers[0].Command
	if len(cmd) < 4 {
		t.Fatalf("unexpected command: %v", cmd)
	}
	if cmd[0] != "crane" || cmd[1] != "copy" {
		t.Fatalf("unexpected command prefix: %v", cmd)
	}
	if cmd[2] != full {
		t.Fatalf("expected src %s in command, got %s", full, cmd[2])
	}
	if cmd[3] != harbor {
		t.Fatalf("expected dst %s in command, got %s", harbor, cmd[3])
	}

	if _, ok := interface{}(job).(*batchv1.Job); !ok {
		t.Fatalf("BuildPullJob did not return *batchv1.Job")
	}
}
