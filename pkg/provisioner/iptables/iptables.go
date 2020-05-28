package iptables

import (
  "log"
  "runtime"
  "github.com/containernetworking/plugins/pkg/ns"
  "github.com/nokia/danm-utils/types/poltypes"
  corev1 "k8s.io/api/core/v1"
  k8stables "k8s.io/kubernetes/pkg/util/iptables"
  "k8s.io/utils/exec"
)

var (
  DefaultInputRules = poltypes.NetRuleChain {
    Name: string(k8stables.ChainInput), Rules: []poltypes.NetRule {
      poltypes.NetRule{SourceIface: "lo", Operation: poltypes.IptablesAccept,},
      poltypes.NetRule{Operation: poltypes.IptablesReject,},
    },
  }
  DefaultOutputRules = poltypes.NetRuleChain {
    Name: string(k8stables.ChainOutput), Rules: []poltypes.NetRule {
      poltypes.NetRule{DestIface: "lo", Operation: poltypes.IptablesAccept,},
      poltypes.NetRule{Operation: poltypes.IptablesReject,},
    },
  }
  DefaultForwardRules = poltypes.NetRuleChain {
    Name: string(k8stables.ChainForward), Rules: []poltypes.NetRule {
      poltypes.NetRule{Operation: poltypes.IptablesReject,},
    },
  }
)

type IptablesProvisioner struct {
  V4Provisioner k8stables.Interface
  V6Provisioner k8stables.Interface
}

func NewIptablesProvisioner() *IptablesProvisioner {
  v4Exec := exec.New()
  v4IptablesClient := k8stables.New(v4Exec, k8stables.ProtocolIPv4)
  v6Exec := exec.New()
  v6IptablesClient := k8stables.New(v6Exec, k8stables.ProtocolIPv6)
  iptablesProv := IptablesProvisioner{V4Provisioner: v4IptablesClient, V6Provisioner: v6IptablesClient}
  return &iptablesProv
}

func (iptabProv *IptablesProvisioner) AddRulesToNewPod(ruleSet *poltypes.NetRuleSet, pod *corev1.Pod) {
  runtime.LockOSThread()
  defer runtime.UnlockOSThread()
  origns, err := ns.GetCurrentNS()
  if err != nil {
    return
  }
  hns, err := ns.GetNS(ruleSet.Netns)
  if err != nil {
    return
  }
  defer func() {
    hns.Close()
    origns.Set()
  }()
  err = ensureChains(iptabProv, ruleSet)
  if err != nil {
    log.Println("required filter chains could not be created for Pod:" + pod.ObjectMeta.Name +
      " in ns:" + pod.ObjectMeta.Namespace + " because of error:" + err.Error())
    return
  }
  provisionDynamicRules(iptabProv, ruleSet, pod)
  provisionDefaultRules(iptabProv, pod)
}

func ensureChains(iptablesProv *IptablesProvisioner, ruleSet *poltypes.NetRuleSet) error {
  err := ensureChain(ruleSet.IngressV4Chain, iptablesProv.V4Provisioner)
  if err != nil {
    return err
  }
  err = ensureChain(ruleSet.IngressV6Chain, iptablesProv.V6Provisioner)
  if err != nil {
    return err
  }
  err = ensureChain(ruleSet.EgressV4Chain, iptablesProv.V4Provisioner)
  if err != nil {
    return err
  }
  return ensureChain(ruleSet.EgressV6Chain, iptablesProv.V6Provisioner)
}

func ensureChain(chain poltypes.NetRuleChain, provisioner k8stables.Interface) error {
  var err error
  if len(chain.Rules) > 0 {
    _, err = provisioner.EnsureChain(k8stables.TableFilter, k8stables.Chain(chain.Name))
  }
  return err
}

func provisionDynamicRules(iptablesProv *IptablesProvisioner, ruleSet *poltypes.NetRuleSet, pod *corev1.Pod) {
  provisionRulesIntoChain(iptablesProv.V4Provisioner, ruleSet.IngressV4Chain, pod)
  provisionRulesIntoChain(iptablesProv.V6Provisioner, ruleSet.IngressV6Chain, pod)
  provisionRulesIntoChain(iptablesProv.V4Provisioner, ruleSet.EgressV4Chain, pod)
  provisionRulesIntoChain(iptablesProv.V6Provisioner, ruleSet.EgressV6Chain, pod)
}

func provisionDefaultRules(iptablesProv *IptablesProvisioner, pod *corev1.Pod) {
  provisionRulesIntoChain(iptablesProv.V4Provisioner, DefaultInputRules, pod)
  provisionRulesIntoChain(iptablesProv.V4Provisioner, DefaultOutputRules, pod)
  provisionRulesIntoChain(iptablesProv.V4Provisioner, DefaultForwardRules, pod)
}

func provisionRulesIntoChain(provisioner k8stables.Interface, rules poltypes.NetRuleChain, pod *corev1.Pod) {
  for _, rule := range rules.Rules {
		args := createArgsFromRule(rule)
		_, err := provisioner.EnsureRule(k8stables.Append, k8stables.TableFilter, k8stables.Chain(rules.Name), args...)
    if err != nil {
      log.Println("ERROR: provisioning iptables rule for Pod: " + pod.ObjectMeta.Name + " in ns: " + pod.ObjectMeta.Namespace + "with args:" + rule.String() +
        " into chain:" + rules.Name + " failed with error:" + err.Error())
    }
  }
}

func createArgsFromRule(rule poltypes.NetRule) []string {
  args := make([]string, 0)
  if rule.Protocol    != "" {args = append(args, "-p", rule.Protocol)}
  if rule.SourcePort  != "" {args = append(args, "--sport", rule.SourcePort)}
  if rule.DestPort    != "" {args = append(args, "--dport", rule.DestPort)}
  if rule.SourceIface != "" {args = append(args, "-i", rule.SourceIface)}
  if rule.DestIface   != "" {args = append(args, "-o", rule.DestIface)}
  if rule.SourceIp    != "" {args = append(args, "-s", rule.SourceIp)}
  if rule.DestIp      != "" {args = append(args, "-d", rule.DestIp)}
  if rule.Operation   != "" {args = append(args, "-j", rule.Operation)
  } else {args = append(args, "-j", poltypes.IptablesAccept)}
  return args
}