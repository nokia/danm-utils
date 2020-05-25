package netruleset

import (
  polv1 "github.com/nokia/danm-utils/crd/api/netpol/v1"
  "github.com/nokia/danm-utils/types/poltypes"
)

func NewNetRuleSet(polSet []polv1.DanmNetworkPolicy, depSet *poltypes.DanmEpBuckets, namespace string) *poltypes.NetRuleSet {
  return &poltypes.NetRuleSet{}
}