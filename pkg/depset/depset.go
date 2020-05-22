package depset

import (
  "log"
  "context"
  danmv1 "github.com/nokia/danm/crd/apis/danm/v1"
  danmclientset "github.com/nokia/danm/crd/client/clientset/versioned"
  "github.com/nokia/danm-utils/types/poltypes"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DanmEpSet struct {
  DanmEps *poltypes.DanmEpBuckets
}

func NewDanmEpSet(danmClient danmclientset.Interface, namespace string) *DanmEpSet {
  var depSet DanmEpSet
  deps, err := danmClient.DanmV1().DanmEps(namespace).List(context.TODO(), metav1.ListOptions{})
  if err != nil {
    log.Println("ERROR: can't list DANM DanmEps API because:" + err.Error())
    return &depSet
  }
  depSet.DanmEps = sortDepsIntoBuckets(deps.Items)
  return &depSet
}

func sortDepsIntoBuckets(deps []danmv1.DanmEp) *poltypes.DanmEpBuckets {
  depBuckets := make(poltypes.DanmEpBuckets, 0)
  depUidCache := make(map[string]poltypes.UidCache, 0)
  for _, dep := range deps {
    for key, value := range dep.ObjectMeta.Labels {
      if _, ok := depUidCache[key+value+poltypes.CustomBucketPostfix][dep.ObjectMeta.UID]; !ok {
        depUidCache[key+value+poltypes.CustomBucketPostfix][dep.ObjectMeta.UID] = true
        depBuckets[key+value+poltypes.CustomBucketPostfix] = append(depBuckets[key+value+poltypes.CustomBucketPostfix], dep)
      }
    }
  }
  return &depBuckets
}