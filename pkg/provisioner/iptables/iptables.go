package iptables

import (
  "github.com/nokia/danm-utils/types/poltypes"
  corev1 "k8s.io/api/core/v1"
  "k8s.io/kubernetes/pkg/util/iptables"
)

type IptablesProvisioner struct {
  V4Provisioner iptables.Interface
  V6Provisioner iptables.Interface
}

func NewIptablesProvisioner() *IptablesProvisioner {
  return &IptablesProvisioner{}
}

func (iptabProv *IptablesProvisioner) AddRulesToPod(ruleSet *poltypes.NetRuleSet, pod *corev1.Pod) error {
  return nil
}