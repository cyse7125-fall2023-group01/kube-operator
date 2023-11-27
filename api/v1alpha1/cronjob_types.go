/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CronJobSpec defines the desired state of CronJob
type CronJobSpec struct {
	ID                     string `json:"ID"`
	Name                   string `json:"name"`
	URI                    string `json:"uri"`
	IsPaused               bool   `json:"is_paused"`
	NumRetries             int    `json:"num_retries"`
	UptimeSLA              int    `json:"uptime_sla"`
	ResponseTimeSLA        int    `json:"response_time_sla"`
	UseSSL                 bool   `json:"use_ssl"`
	ResponseStatusCode     int    `json:"response_status_code"`
	CheckIntervalInSeconds int    `json:"check_interval_in_seconds"`
}

// CronJobStatus defines the observed state of CronJob
type CronJobStatus struct {
	LastExecutionTime metav1.Time `json:"last_execution_time,omitempty"`
	Status            string      `json:"status,omitempty"`
	Success           bool        `json:"success,omitempty"`
	ErrorMessage      string      `json:"errorMessage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CronJob is the Schema for the cronjobs API
type CronJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronJobSpec   `json:"spec,omitempty"`
	Status CronJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CronJobList contains a list of CronJob
type CronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronJob{}, &CronJobList{})
}
