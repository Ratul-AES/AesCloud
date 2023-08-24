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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CorePodSpec defines the desired state of CorePod
type CorePodSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CorePod. Edit corepod_types.go to remove/update
	Size           int    `json:"size,omitempty"`
	PvSize         int    `json:"pvSize,omitempty"`
	DbImg          string `json:"dbImg,omitempty"`
	OrgImg         string `json:"coreImg,omitempty"`
	FrontendImg    string `json:"frontendImg,omitempty"`
	FrontendImgFs  string `json:"frontendImgFs,omitempty"`
	FrontendImgOrg string `json:"frontendImgOrg,omitempty"`
	FrontendImgUm  string `json:"frontendImgUm,omitempty"`
	GoImg          string `json:"goImg,omitempty"`
	BeReplicas     int32  `json:"beReplicas,omitempty"`
}

// CorePodStatus defines the observed state of CorePod
type CorePodStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name     string `json:"name,omitempty"`
	Progress string `json:"progress,omitempty"`
	Ready    string `json:"ready,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CorePod is the Schema for the corepods API
type CorePod struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CorePodSpec   `json:"spec,omitempty"`
	Status CorePodStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CorePodList contains a list of CorePod
type CorePodList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CorePod `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CorePod{}, &CorePodList{})
}
