package v1

import (
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/types"
  "k8s.io/pkg/apis/networking"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DanmNetworkPolicy struct {
  meta_v1.TypeMeta   `json:",inline"`
  meta_v1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
  Spec NetPolSpec    `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

type NetPolSpec struct {
  PodSelector meta_v1.LabelSelector    `json:"podSelector" protobuf:"bytes,1,opt,name=podSelector"`
  Egress      NetworkPolicyEgressRule  `json:"ingress,omitempty" protobuf:"bytes,2,rep,name=ingress"`
  Ingress     NetworkPolicyIngressRule `json:"egress,omitempty" protobuf:"bytes,3,rep,name=egress"`
  PolicyTypes []networking.PolicyType  `json:"policyTypes,omitempty" protobuf:"bytes,4,rep,name=policyTypes"`
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
	PodSelector       meta_v1.LabelSelector `json:"podSelector,omitempty" protobuf:"bytes,1,opt,name=podSelector"`
	NamespaceSelector meta_v1.LabelSelector `json:"namespaceSelector,omitempty" protobuf:"bytes,2,opt,name=namespaceSelector"`
  NetworkSelector   NetworkSelector       `json:"networkSelector,omitempty" protobuf:"bytes,3,opt,name=networkSelector"`
}

type NetworkSelector struct {
  Name  string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
  Type  string `json:"type,omitempty" protobuf:"bytes,2,opt,name=type"`
}