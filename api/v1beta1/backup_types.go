/*
Copyright 2024.

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

// Datasource: MySQL
type Origin struct {
	// The address of the database
	Host string `json:"host"`
	// The port
	Port int32 `json:"port"`
	// Username for the batabase
	Username string `json:"username"`
	// Password for the origin database, should be in the Secret
	Password string `json:"password"`
}

// Endpoint: minIO
type Destination struct {
	// The address of the object storage system
	Endpoint string `json:"endpoint"`
	// The accesskey of the object storage system
	AccessKey string `json:"accessKey"`
	// The accessSecret of the object storage system
	AccessSecret string `json:"accessSecret"`
	// The bucketname of the object storage system
	BucketName string `json:"bucketName"`
}

// BackUpSpec defines the desired state of BackUp
type BackUpSpec struct {
	// Whether the backup is enabled
	Enable bool `json:"enable"`
	// The expected time when the backup starts, example: 11:30
	StartTime string `json:"startTime"`
	// The period for backups, in minutes.
	Period int `json:"period"`
	// The Origin of the data source
	Origin Origin `json:"origin"`
	// The destination of the backup
	Destination Destination `json:"destination"`
}

// BackUpStatus defines the observed state of BackUp
type BackUpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Whether the backup is active
	Active bool `json:"active"`
	// The time when the next backup will start
	NextTime int64 `json:"nextTime"`
	//The result of the last backup
	LastBackupResult string `json:"lastBackupResult"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// BackUp is the Schema for the backups API
type BackUp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackUpSpec   `json:"spec,omitempty"`
	Status BackUpStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackUpList contains a list of BackUp
type BackUpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackUp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackUp{}, &BackUpList{})
}
