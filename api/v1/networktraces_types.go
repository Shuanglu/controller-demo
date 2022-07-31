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

// NetworktracesSpec defines the desired state of Networktraces
type NetworktracesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Networktraces. Edit networktraces_types.go to remove/update
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

// NetworktracesStatus defines the observed state of Networktraces
type NetworktracesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// whether the captrue operation completes
	Completed string `json:"completed"`
	// count of successful capture
	Successful string `json:"successful"`
	// pod name of the failed capture and reason of the failure
	Failed map[string]string `json:"failed"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Networktraces is the Schema for the networktraces API
type Networktraces struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworktracesSpec   `json:"spec,omitempty"`
	Status NetworktracesStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NetworktracesList contains a list of Networktraces
type NetworktracesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Networktraces `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Networktraces{}, &NetworktracesList{})
}
