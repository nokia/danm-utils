package polset

import (
  "log"
  "context"
  polv1 "github.com/nokia/danm-utils/crd/api/netpol/v1"
  polclientset "github.com/nokia/danm-utils/crd/client/clientset/versioned"
  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/types"
)

const (
  DefaultBucketName   = "default"
  CustomBucketPostfix = "bucket42"
)

type PolicySet struct {
  NetPols map[string][]polv1.DanmNetworkPolicy
}

func NewPolicySet(netpolClient polclientset.Interface, namespace string) *PolicySet {
  var polSet PolicySet
  polSet.NetPols = make(map[string][]polv1.DanmNetworkPolicy, 0)
  netPols, err := netpolClient.NetpolV1().DanmNetworkPolicies(namespace).List(context.TODO(), metav1.ListOptions{})
  if err != nil {
    log.Println("ERROR: can't list DANM NetworkPolicies API because:" + err.Error())
    return &polSet
  }
  polSet.NetPols = sortPoliciesIntoBuckets(netPols.Items)
  return &polSet
}

func sortPoliciesIntoBuckets(netPols []polv1.DanmNetworkPolicy) map[string][]polv1.DanmNetworkPolicy {
  polBuckets := make(map[string][]polv1.DanmNetworkPolicy, 0)
  for _, policy := range netPols {
    selectors, err := metav1.LabelSelectorAsMap(&policy.Spec.PodSelector)
    if err != nil {
      log.Println("WARNING: PodSelector field of DanmNetworkPolicy:" + policy.ObjectMeta.Name + " in namespace:"
        + policy.ObjectMeta.Namespace + " could not be parsed and is therefore ignored because of error:" + err.Error())
      continue
    }
    if len(selectors) == 0 {
      //From K8s documentation: "an empty podSelector selects all pods in the namespace"
      polBuckets[DefaultBucketName] = append(polBuckets[DefaultBucketName], policy)
    } else {
      for key, value := range selectors {
        polBuckets[key+value+CustomBucketPostfix] = append(polBuckets[key+value+CustomBucketPostfix], policy)
      }
    }
  }
  return polBuckets
}

func (polSet *PolicySet) FilterApplicablePolicies(pod *corev1.Pod) []polv1.DanmNetworkPolicy {
  polUidCache := make(map[types.UID]bool, 0)
  applicablePolicies := make([]polv1.DanmNetworkPolicy, 0)
  for key, value := range pod.ObjectMeta.Labels {
    //there are NetworkPolicies selecting the Pod cause a bucket for this specific label exists
    if policies, ok := polSet.NetPols[key+value+CustomBucketPostfix]; ok {
      applicablePolicies, polUidCache = filterPoliciesWithoutDupes(policies, applicablePolicies, polUidCache)
    }
  }
  //Default policies are selecting all Pods in the namespace, so if the default bucket exists we need to treat all pols in it as applicable
  if policies, ok := polSet.NetPols[DefaultBucketName]; ok {
    applicablePolicies, polUidCache = filterPoliciesWithoutDupes(policies, applicablePolicies, polUidCache)
  }
  return applicablePolicies
}

func filterPoliciesWithoutDupes(policies, applicablePolicies []polv1.DanmNetworkPolicy, podUidCache map[types.UID]bool) ([]polv1.DanmNetworkPolicy,map[types.UID]bool){
  for _, policy := range policies {
    //the same policy might select the Pod multiple times via different labels
    //we need to weed out the duplicates by not adding them again if they are already in the list
    if _, ok := podUidCache[policy.ObjectMeta.UID]; !ok {
      podUidCache[policy.ObjectMeta.UID] = true
      applicablePolicies = append(applicablePolicies, policy)
    }
  }
  return applicablePolicies, podUidCache
}