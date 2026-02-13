package configfile

import (
	"testing"

	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/domain/resource"
)

func TestFilterResourcesBySubmitTypeJob(t *testing.T) {
	resources := []resource.Resource{
		{Type: resource.ResourceJob},
		{Type: resource.ResourceConfigMap},
		{Type: resource.ResourceService},
		{Type: resource.ResourceType("workflow")},
	}

	filtered, err := filterResourcesBySubmitType(resources, string(executor.SubmitTypeJob))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(filtered) != 3 {
		t.Fatalf("expected 3 resources, got %d", len(filtered))
	}
	for _, res := range filtered {
		if string(res.Type) == "workflow" {
			t.Fatalf("unexpected workflow resource in job submit")
		}
	}
}

func TestFilterResourcesBySubmitTypeWorkflow(t *testing.T) {
	resources := []resource.Resource{
		{Type: resource.ResourceJob},
		{Type: resource.ResourceConfigMap},
		{Type: resource.ResourceType("workflow")},
	}

	filtered, err := filterResourcesBySubmitType(resources, string(executor.SubmitTypeWorkflow))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(filtered))
	}
	for _, res := range filtered {
		if string(res.Type) == "job" {
			t.Fatalf("unexpected job resource in workflow submit")
		}
	}
}

func TestFilterResourcesBySubmitTypeInvalid(t *testing.T) {
	resources := []resource.Resource{{Type: resource.ResourceJob}}
	_, err := filterResourcesBySubmitType(resources, "nope")
	if err == nil {
		t.Fatalf("expected error for invalid submit type")
	}
}

func TestFilterResourcesBySubmitTypeMissingWorkload(t *testing.T) {
	resources := []resource.Resource{{Type: resource.ResourceConfigMap}}
	_, err := filterResourcesBySubmitType(resources, string(executor.SubmitTypeJob))
	if err == nil {
		t.Fatalf("expected error when no workload resources are present")
	}
}
