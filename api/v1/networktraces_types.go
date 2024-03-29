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

// NetworktraceSpec defines the desired state of Networktrace
type NetworktraceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Networktrace. Edit Networktrace_types.go to remove/update
	//Foo string `json:"foo,omitempty"`
	// "kind" of the target resource. Allowed values are: 'deployment'/'daemonset'/'statefulset'/'pod'
	Kind string `json:"kind"`
	// "name" of the target resource
	Name string `json:"name"`
	// "namespace" of the target resource
	Namespace string `json:"namespace"`
	// "duration" of the capture
	Duration string `json:"duration"`
	// "labels selector" of the target resource
	Labels map[string]string `json:"labels,omitempty"`
}

// NetworktraceStatus defines the observed state of Networktrace
type NetworktraceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// whether the captrue operation completes
	Completed string `json:"completed"`
	// count of successful capture
	Successful string `json:"successful"`
	// pod name of the failed capture and reason of the failure
	Failed map[string]string `json:"failed"`
	// pod name of the active pods used for capture
	Active []string `json:"active"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Networktrace is the Schema for the Networktrace API
type Networktrace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworktraceSpec   `json:"spec,omitempty"`
	Status NetworktraceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetworktraceList contains a list of Networktrace
type NetworktraceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Networktrace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Networktrace{}, &NetworktraceList{})
}
