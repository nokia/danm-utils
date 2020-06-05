package v1

import (
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/kubernetes/pkg/apis/networking"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DanmNetworkPolicy struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
  Spec NetPolSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DanmNetworkPolicyList struct {
  metav1.TypeMeta `json:",inline"`
  metav1.ListMeta `json:"metadata"`
  Items           []DanmNetworkPolicy `json:"items"`
}

type NetPolSpec struct {
  PodSelector  metav1.LabelSelector       `json:"podSelector" protobuf:"bytes,1,opt,name=podSelector"`
  Ingress     []NetworkPolicyIngressRule  `json:"ingress,omitempty" protobuf:"bytes,2,rep,name=ingress"`
  Egress      []NetworkPolicyEgressRule   `json:"egress,omitempty" protobuf:"bytes,3,rep,name=egress"`
  PolicyTypes []networking.PolicyType     `json:"policyTypes,omitempty" protobuf:"bytes,4,rep,name=policyTypes"`
}

type NetworkPolicyIngressRule struct {
  Ports []networking.NetworkPolicyPort `json:"ports,omitempty" protobuf:"bytes,1,rep,name=ports"`
  From  []NetworkPolicyPeer            `json:"from,omitempty" protobuf:"bytes,2,rep,name=from"`
}

type NetworkPolicyEgressRule struct {
  Ports []networking.NetworkPolicyPort `json:"ports,omitempty" protobuf:"bytes,1,rep,name=ports"`
  To    []NetworkPolicyPeer            `json:"to,omitempty" protobuf:"bytes,2,rep,name=to"`
}

type NetworkPolicyPeer struct {
  PodSelector       metav1.LabelSelector `json:"podSelector,omitempty" protobuf:"bytes,1,opt,name=podSelector"`
  NamespaceSelector metav1.LabelSelector `json:"namespaceSelector,omitempty" protobuf:"bytes,2,opt,name=namespaceSelector"`
  NetworkSelector   NetworkSelector      `json:"networkSelector,omitempty" protobuf:"bytes,3,opt,name=networkSelector"`
}

type NetworkSelector struct {
  Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
  Type  string `json:"type,omitempty" protobuf:"bytes,2,opt,name=type"`
}