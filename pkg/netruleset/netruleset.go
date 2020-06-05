package netruleset

import (
  "log"
  "strings"
  "github.com/nokia/danm/pkg/ipam"
  polv1 "github.com/nokia/danm-utils/crd/api/netpol/v1"
  "github.com/nokia/danm-utils/types/poltypes"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/kubernetes/pkg/apis/networking"
)

type RuleParser func(address string, ports []networking.NetworkPolicyPort) []poltypes.NetRule

func NewNetRuleSet(polSet []polv1.DanmNetworkPolicy, depSet poltypes.DanmEpBuckets, netns string) *poltypes.NetRuleSet {
  ruleSet := poltypes.NetRuleSet{Netns: netns}
  ruleSet.IngressV4Chain.Name = poltypes.IngressV4ChainName
  ruleSet.IngressV6Chain.Name = poltypes.IngressV6ChainName
  ruleSet.EgressV4Chain.Name = poltypes.EgressV4ChainName
  ruleSet.EgressV6Chain.Name = poltypes.EgressV6ChainName
  for _, policy := range polSet {
    //TODO: is it really necessary for Ingress / Egress to be list?
    // Format is kept to be consistent with upstream, but there is really no use-case for having multiple from/to sections in a network policy
    if len(policy.Spec.Ingress) > 0 {
      ruleSet.IngressV4Chain.Rules, ruleSet.IngressV6Chain.Rules = parseAndAppendPolicyRules(depSet, policy.Spec.Ingress[0].From, policy.Spec.Ingress[0].Ports, newIngressNetRules)
    }
    if len(policy.Spec.Egress) > 0 {
      ruleSet.EgressV4Chain.Rules, ruleSet.EgressV6Chain.Rules = parseAndAppendPolicyRules(depSet, policy.Spec.Egress[0].To, policy.Spec.Egress[0].Ports, newEgressNetRules)
    }
  }
  return &ruleSet
}

func parseAndAppendPolicyRules(depSet poltypes.DanmEpBuckets, peers []polv1.NetworkPolicyPeer, ports []networking.NetworkPolicyPort, parserFunc RuleParser) ([]poltypes.NetRule,[]poltypes.NetRule) {
  v4Rules := make([]poltypes.NetRule, 0)
  v6Rules := make([]poltypes.NetRule, 0)
  for _, peer := range peers {
    depCache := make(poltypes.UidCache, 0)
    selectors, err := metav1.LabelSelectorAsMap(&peer.PodSelector)
    if err != nil {
      log.Println("WARNING: PodSelector parsing failed with error:" + err.Error() + ", ignoring related peers!")
      continue
    }
    for key, value := range selectors {
      for _, dep := range depSet[key+value+poltypes.CustomBucketPostfix] {
        if _, ok := depCache[dep.ObjectMeta.UID]; !ok {
          depCache[dep.ObjectMeta.UID] = true
          if dep.Spec.Iface.Address != "" && dep.Spec.Iface.Address != ipam.NoneAllocType {
            v4Rules = append(v4Rules, parserFunc(dep.Spec.Iface.Address, ports)...)
          }
          if dep.Spec.Iface.AddressIPv6 != "" && dep.Spec.Iface.AddressIPv6 != ipam.NoneAllocType {
            v6Rules = append(v6Rules, parserFunc(dep.Spec.Iface.AddressIPv6, ports)...)
          }
        }
      }
    }
  }
  return v4Rules, v6Rules
}

func newIngressNetRules(address string, ports []networking.NetworkPolicyPort) []poltypes.NetRule {
  ingressRules := make([]poltypes.NetRule, 0)
  if len(ports) == 0 {
    universalRule := poltypes.NetRule{SourceIp: strings.Split(address, "/")[0]}
    ingressRules = append(ingressRules, universalRule)
    return ingressRules
  }
  for _, port := range ports {
    ingressRule := poltypes.NetRule{SourceIp: strings.Split(address, "/")[0], SourcePort: port.Port.StrVal, Protocol: string(*port.Protocol)}
    ingressRules = append(ingressRules, ingressRule)
  }
  return ingressRules
}

func newEgressNetRules(address string, ports []networking.NetworkPolicyPort) []poltypes.NetRule {
  egressRules := make([]poltypes.NetRule, 0)
  if len(ports) == 0 {
    universalRule := poltypes.NetRule{DestIp: strings.Split(address, "/")[0]}
    egressRules = append(egressRules, universalRule)
    return egressRules
  }
  for _, port := range ports {
    egressRule := poltypes.NetRule{DestIp: strings.Split(address, "/")[0], DestPort: port.Port.StrVal, Protocol: string(*port.Protocol)}
    egressRules = append(egressRules, egressRule)
  }
  return egressRules
}