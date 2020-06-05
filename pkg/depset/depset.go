package depset

import (
  "context"
  "log"
  danmv1 "github.com/nokia/danm/crd/apis/danm/v1"
  danmclientset "github.com/nokia/danm/crd/client/clientset/versioned"
  "github.com/nokia/danm-utils/types/poltypes"
  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DanmEpSet struct {
  DanmEps *poltypes.DanmEpBuckets
  PodEps  []danmv1.DanmEp
}

func NewDanmEpSet(danmClient danmclientset.Interface, pod *corev1.Pod) *DanmEpSet {
  var depSet DanmEpSet
  deps, err := danmClient.DanmV1().DanmEps(pod.ObjectMeta.Namespace).List(context.TODO(), metav1.ListOptions{})
  if err != nil {
    log.Println("ERROR: can't list DANM DanmEps API because:" + err.Error())
    return &depSet
  }
  depSet.DanmEps, depSet.PodEps = sortDeps(deps.Items, pod)
  return &depSet
}

func sortDeps(deps []danmv1.DanmEp, pod *corev1.Pod) (*poltypes.DanmEpBuckets,[]danmv1.DanmEp)  {
  depBuckets := make(poltypes.DanmEpBuckets, 0)
  depUidCache := make(map[string]poltypes.UidCache, 0)
  podEps := make([]danmv1.DanmEp, 0)
  for _, dep := range deps {
    if dep.Spec.PodUID == pod.ObjectMeta.UID {
      podEps = append(podEps, dep)
      continue
    }
    for key, value := range dep.ObjectMeta.Labels {
      if _, ok := depUidCache[key+value+poltypes.CustomBucketPostfix][dep.ObjectMeta.UID]; !ok {
        if depUidCache[key+value+poltypes.CustomBucketPostfix] == nil {
          cache := make(poltypes.UidCache, 0)
          depUidCache[key+value+poltypes.CustomBucketPostfix] = cache
        }
        depUidCache[key+value+poltypes.CustomBucketPostfix][dep.ObjectMeta.UID] = true
        depBuckets[key+value+poltypes.CustomBucketPostfix] = append(depBuckets[key+value+poltypes.CustomBucketPostfix], dep)
      }
    }
  }
  return &depBuckets, podEps
}