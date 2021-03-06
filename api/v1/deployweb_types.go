/*
Copyright 2022.

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

// DeploywebSpec defines the desired state of Deployweb
type DeploywebSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// 镜像
	Image string `json:"image"`
	// 端口
	Port int32 `json:"port"`
	//
	Replicas *int32 `json:"replicas"`
}

// DeploywebStatus defines the observed state of Deployweb
type DeploywebStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Deployweb is the Schema for the deploywebs API
type Deployweb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploywebSpec   `json:"spec,omitempty"`
	Status DeploywebStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeploywebList contains a list of Deployweb
type DeploywebList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Deployweb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Deployweb{}, &DeploywebList{})
}
