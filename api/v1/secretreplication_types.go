/*

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

// SecretReplicationSpec defines the desired state of SecretReplication
type SecretReplicationSpec struct {
	// Source namespace the secret lives in
	SrcNamespace string `json:"srcNamespace"`
	// Name of the source secret to replicate
	SrcName string `json:"srcName"`
	// Destination namespace of the secret to replicate
	DstNamespace string `json:"dstNamespace"`
	// Name of the replicated secret in the destination namespace
	DstName string `json:"dstName"`
}

// SecretReplicationStatus defines the observed state of SecretReplication
type SecretReplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// SecretReplication is the Schema for the secretreplications API
type SecretReplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretReplicationSpec   `json:"spec,omitempty"`
	Status SecretReplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretReplicationList contains a list of SecretReplication
type SecretReplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretReplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecretReplication{}, &SecretReplicationList{})
}
