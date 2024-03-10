package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Peer struct {
	Id        int32  `json:"id"`
	PublicKey string `json:"publicKey"`
	Endpoint  string `json:"endpoints"`
}

// ConnectSpec defines the desired state of Connect
type ConnectSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=255
	Id int32 `json:"id"`

	// later will store to secrets
	PrivateKey *string `json:"privateKey,omitempty"`
	PublicKey  *string `json:"publicKey,omitempty"`
	Ip         *string `json:"ip,omitempty"`

	Endpoint string `json:"endpoints,omitempty"`

	Peers []Peer `json:"peers,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Connect is the Schema for the connects API
type Connect struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (d *Connect) EnsureGVK() {
	if d != nil {
		d.SetGroupVersionKind(GroupVersion.WithKind("Connect"))
	}
}

func (d *Connect) GetStatus() *rApi.Status {
	return &d.Status
}

func (d *Connect) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (d *Connect) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: GroupVersion.WithKind("Connect").String(),
	}
}

//+kubebuilder:object:root=true

// ConnectList contains a list of Connect
type ConnectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Connect `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Connect{}, &ConnectList{})
}
