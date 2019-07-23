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

// MinionSpec defines the desired state of Minion
type MinionSpec struct {
}

// MinionStatus defines the observed state of Minion
type MinionStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=minions,scope=Cluster
// +kubebuilder:subresource:status

// Minion is the Schema for the minions API
type Minion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MinionSpec   `json:"spec,omitempty"`
	Status MinionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MinionList contains a list of Minion
type MinionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Minion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Minion{}, &MinionList{})
}
