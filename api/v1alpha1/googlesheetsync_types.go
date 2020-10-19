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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GooglesheetSyncSpec defines the desired state of GooglesheetSync
type GooglesheetSyncSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of GooglesheetSync. Edit GooglesheetSync_types.go to remove/update
	SpreadsheetID string `json:"spreadsheet_id,omitempty"`
	CellRange     string `json:"cell_range,omitempty"`
}

// GooglesheetSyncStatus defines the observed state of GooglesheetSync
type GooglesheetSyncStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SyncStatus string `json:"sync_status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GooglesheetSync is the Schema for the googlesheetsyncs API
type GooglesheetSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GooglesheetSyncSpec   `json:"spec,omitempty"`
	Status GooglesheetSyncStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GooglesheetSyncList contains a list of GooglesheetSync
type GooglesheetSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GooglesheetSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GooglesheetSync{}, &GooglesheetSyncList{})
}
