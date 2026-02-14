package configfile

func injectJobLabels(obj map[string]interface{}, jobID, commitID string) {
	if obj == nil || jobID == "" {
		return
	}

	setLabels := func(metadata map[string]interface{}) {
		if metadata == nil {
			return
		}
		labels, ok := metadata["labels"].(map[string]interface{})
		if !ok {
			labels = make(map[string]interface{})
			metadata["labels"] = labels
		}
		labels["platform.job-id"] = jobID
		if commitID != "" {
			labels["platform.configcommit-id"] = commitID
		}
	}

	if metadata, ok := obj["metadata"].(map[string]interface{}); ok {
		setLabels(metadata)
	}

	spec, _ := obj["spec"].(map[string]interface{})
	if spec == nil {
		return
	}

	if template, ok := spec["template"].(map[string]interface{}); ok {
		if metadata, ok := template["metadata"].(map[string]interface{}); ok {
			setLabels(metadata)
		}
	}

	if jobTemplate, ok := spec["jobTemplate"].(map[string]interface{}); ok {
		if jobSpec, ok := jobTemplate["spec"].(map[string]interface{}); ok {
			if template, ok := jobSpec["template"].(map[string]interface{}); ok {
				if metadata, ok := template["metadata"].(map[string]interface{}); ok {
					setLabels(metadata)
				}
			}
		}
	}
}
