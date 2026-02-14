package configfile

import (
	"fmt"
	"strings"

	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/domain/resource"
)

func filterResourcesBySubmitType(resources []resource.Resource, submitType string) ([]resource.Resource, error) {
	if submitType == "" {
		return resources, nil
	}
	mode := strings.ToLower(submitType)
	if mode != string(executor.SubmitTypeJob) && mode != string(executor.SubmitTypeWorkflow) {
		return nil, fmt.Errorf("invalid submit_type: %s", submitType)
	}

	filtered := make([]resource.Resource, 0, len(resources))
	workloadCount := 0
	for _, res := range resources {
		kind := strings.ToLower(string(res.Type))
		if kind == "" {
			filtered = append(filtered, res)
			continue
		}
		if isJobWorkloadKind(kind) {
			if mode == string(executor.SubmitTypeJob) {
				filtered = append(filtered, res)
				workloadCount++
			}
			continue
		}
		if isWorkflowWorkloadKind(kind) {
			if mode == string(executor.SubmitTypeWorkflow) {
				filtered = append(filtered, res)
				workloadCount++
			}
			continue
		}
		filtered = append(filtered, res)
	}
	if workloadCount == 0 {
		return nil, fmt.Errorf("no %s workload resources found in configfile", mode)
	}
	return filtered, nil
}

func isJobWorkloadKind(kind string) bool {
	switch kind {
	case "job", "cronjob", "flashjob":
		return true
	default:
		return false
	}
}

func isWorkflowWorkloadKind(kind string) bool {
	switch kind {
	case "workflow", "workflowtemplate", "cronworkflow":
		return true
	default:
		return false
	}
}
