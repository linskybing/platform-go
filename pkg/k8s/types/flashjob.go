package types

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FlashJobSpec defines the desired state of FlashJob.
type FlashJobSpec struct {
	MinAvailable int32                `json:"minAvailable,omitempty"`
	Tasks        []v1.PodTemplateSpec `json:"tasks,omitempty"`
}

// FlashJobPhase is a label for the overall status of a FlashJob.
type FlashJobPhase string

const (
	FlashJobPending   FlashJobPhase = "Pending"
	FlashJobRunning   FlashJobPhase = "Running"
	FlashJobSucceeded FlashJobPhase = "Succeeded"
	FlashJobFailed    FlashJobPhase = "Failed"
)

// FlashJobStatus defines the observed state of FlashJob.
type FlashJobStatus struct {
	Active     int32              `json:"active"`
	Succeeded  int32              `json:"succeeded"`
	Failed     int32              `json:"failed"`
	Desired    int32              `json:"desired"`
	Phase      FlashJobPhase      `json:"phase,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// FlashJob is the Schema for the flashjobs API.
type FlashJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlashJobSpec   `json:"spec,omitempty"`
	Status FlashJobStatus `json:"status,omitempty"`
}

// FlashJobList contains a list of FlashJob.
type FlashJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlashJob `json:"items"`
}
