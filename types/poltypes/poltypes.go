package poltypes

import (
  "net"
  danmv1 "github.com/nokia/danm/crd/apis/danm/v1"
  corev1 "k8s.io/api/core/v1"
  "k8s.io/apimachinery/pkg/types"
)

const (
  DefaultBucketName   = "default"
  CustomBucketPostfix = "bucket42"
)

type RuleProvisioner interface {
  AddRulesToPod(*NetRuleSet,*corev1.Pod) error
}

type UidCache map[types.UID]bool

type DanmEpBuckets map[string][]danmv1.DanmEp

type NetRuleSet struct {
  IngressV4Chain NetRuleChain
  IngressV6Chain NetRuleChain
  EgressV4Chain  NetRuleChain
  EgressV6Chain  NetRuleChain
}

type NetRuleChain struct {
  Name string
  Rules []NetRule
}

type NetRule struct {
  SourceIp   *net.IPNet
  SourcePort int
  DestIp     *net.IPNet
  DestPort   int
  Protocol   string
}