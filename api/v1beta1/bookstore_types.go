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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BookstoreSpec defines the desired state of Bookstore
type BookstoreSpec struct {
	// CustomerId is the ID of the customer this Bookstore is associated with.
	CustomerId string `json:"customerId"`
}

// BookstoreStatus defines the observed state of Bookstore
type BookstoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Bookstore is the Schema for the bookstores API
type Bookstore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookstoreSpec   `json:"spec,omitempty"`
	Status BookstoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BookstoreList contains a list of Bookstore
type BookstoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bookstore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bookstore{}, &BookstoreList{})
}
