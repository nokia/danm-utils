package cleaner

import (
  "fmt"
  "time"
  "github.com/nokia/danm/pkg/ipam"
  danmv1 "github.com/nokia/danm/pkg/crd/apis/danm/v1"
  danmclientset "github.com/nokia/danm/pkg/crd/client/clientset/versioned"
  danmscheme "github.com/nokia/danm/pkg/crd/client/clientset/versioned/scheme"
  danminformers "github.com/nokia/danm/pkg/crd/client/informers/externalversions/danm/v1"
  danmlisters "github.com/nokia/danm/pkg/crd/client/listers/danm/v1"
  corev1 "k8s.io/api/core/v1"
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/api/errors"
  "k8s.io/apimachinery/pkg/labels"
  "k8s.io/apimachinery/pkg/util/runtime"
  "k8s.io/apimachinery/pkg/util/wait"
  coreinformers "k8s.io/client-go/informers/core/v1"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/kubernetes/scheme"
  corelisters "k8s.io/client-go/listers/core/v1"
  "k8s.io/client-go/tools/cache"
  "k8s.io/client-go/util/workqueue"
)

type Cleaner struct {
  danmClient    danmclientset.Interface
  initialized   bool
  podLister     corelisters.PodLister
  podSynced     cache.InformerSynced
  workqueue     workqueue.RateLimitingInterface
}

func New(
  danmClient danmclientset.Interface,
  podInformer coreinformers.PodInformer) *Cleaner {
  danmscheme.AddToScheme(scheme.Scheme)
  cleaner := &Cleaner{
    danmClient:    danmClient,
    initialized:   false,
    podLister:     podInformer.Lister(),
    podSynced:     podInformer.Informer().HasSynced,
    workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Endpoints"),
  }
  podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
    UpdateFunc: cleaner.updatePod,
    DeleteFunc: cleaner.delPod,
  })
  return cleaner
}

func (c *Cleaner) Initialize() bool {
  interval := 100
  timeout := 10000
  for timer := 0; timer <= timeout; timer = timer + interval {
    //This can be easily expanded if we need to add more Informers to Cleaner in the future
    if c.podSynced() {
      break
    }
    time.Sleep(interval * time.Millisecond)
  }
  c.initialized = true
  return true
}

func PeriodicCleanup(danmClient danmclientset.Interface, podLister corelisters.PodLister, stopCh <-chan struct{}) {
  go cleanupOnTick(danmClient, podLister)
  log.Println("INFO: Successfully started Cleaner's periodic worker thread")
  <-stopCh
  log.Println("INFO: Shutting down Cleaner's periodic worker thread")
}

func cleanupOnTick(danmClient danmclientset.Interface, podLister corelisters.PodLister) {
  timeForCleanup := time.NewTicker(10 * time.Second)
  for {
    select {
    case <-timeForCleanup.C:
      danmeps, err := danmep.FindByPodName(c.danmClient, "", "")
      if err != nil {
        log.Println("WARNING: Periodic cleaning failed with error:" + err.Error())
        continue
      }
      cleanDanglingEps(danmeps, podLister)
    }
  }
}

func cleanDanglingEps(danmeps []danmv1.DanmEp, podLister corelisters.PodLister) {
  podCache := make(map[string]bool, 0)
  for _, dep := range danmeps {
    //We have already checked this Pod
    if doesPodExist, ok := podCache[dep.ObjectMeta.Namespace+dep.Spec.Pod]; ok {
      if !doesPodExist {
        deleteInterface(dep)
      }
      continue
    }
    _, err := podLister.Pods(dep.ObjectMeta.Namespace).Get(dep.Spec.Pod)
    if errors.IsNotFound(err) {
      deleteInterface(dep)
      podCache[dep.ObjectMeta.Namespace+dep.Spec.Pod] = false
    } else {
      podCache[dep.ObjectMeta.Namespace+dep.Spec.Pod] = true
    }
  }
}

func (c *Cleaner) Run(threadiness int, stopCh <-chan struct{}) error {
  defer runtime.HandleCrash()
  defer c.workqueue.ShutDown()
  log.Println("INFO: starting Cleaner")
  log.Println("INFO: waiting for Cleaner to synchronize cache")
  if ok := cache.WaitForCacheSync(stopCh, c.podSynced); !ok {
    return errors.New("synching Cleaner's cache failed")
  }
  log.Println("INFO: Starting Cleaner event handler threads")
  for i := 0; i < threadiness; i++ {
    go wait.Until(c.runWorker, time.Second, stopCh)
  }
  log.Println("INFO: Successfully started Cleaner's event handler threads")
  <-stopCh
  log.Println("INFO: Shutting down Cleaner's event handler threads")
  return nil
}

func (c *Cleaner) runWorker() {
  for c.processNextWorkItem() {}
}

func (c *Cleaner) processNextWorkItem() bool {
  obj, shutdown := c.workqueue.Get()
  if shutdown {
    return false
  }
  err := c.processItemInQueue(obj)
  if err != nil {
    runtime.HandleError(err)
  }
  return true
}

func (c *Cleaner) processItemInQueue(obj interface{}) error {
  defer c.workqueue.Done(obj)
  var key string
  var ok bool
  if key, ok = obj.(string); !ok {
    c.workqueue.Forget(obj)
    runtime.HandleError(fmt.Errorf("WARNING: Cannot decode work item from queue because instead string type we got %#v", obj))
    return nil
  }
  if err := c.handleKey(key); err != nil {
    return errors.New("ERROR: could not process item:" + key + " because:" + err.Error())
  }
  c.workqueue.Forget(obj)
  return nil
}

func (c *Cleaner) handleKey(key string) error {
  ns, name, err := cache.SplitMetaNamespaceKey(key)
  if err != nil {
    log.Println("WARNING: Dropping work item from because its key:" + key + " could not be broken up into API object identifiers due to error:" + err.Error())
    return nil
  }
  //We give time for DANM to execute normal CNI DEL operation
  //We want to avoid possible interference, and with it exotic race conditions
  //TODO: this quite possibly needs to be more sophisticated than this :)  
  time.Sleep(1 * time.Second)
  deps, err := danmep.FindByPodName(c.danmClient, name, ns)
  if err != nil {
    return err
  }
  //Check if the specified DanmEp (if any) actually exists in the namespace
  for _, dep := range deps {
    log.Println("INFO: Cleaner freeing IPs belonging to interface:" + dep.Spec.Iface.Name + " in Pod:" + dep.Spec.Pod)
    c.deleteInterface(dep)
  }
  return nil
}

func (c *Cleaner) deleteInterface(ep *danmv1.DanmEp) {
  netInfo, err := netcontrol.GetNetworkFromEp(c.danmclient, ep)
  if err != nil {
    log.Println("WARNING: Danmep:" + ep.MetaData.Name + " in namespace:" + ep.MetaData.Namespace + "could not be cleaned as its network could not be GET from K8s API server:" + err.Error())
    return
  }
  //TODO: this definitely need to be expanded into a framework, where network type specific cleanup operations can be plugged-in
  ipam.GarbageCollectIps(c.danmclient, *netInfo, ep.Spec.Iface.Address, ep.Spec.Iface.AddressIPv6)
  c.danmclient.DanmV1().DanmEps(ep.ObjectMeta.Namespace).Delete(ep.ObjectMeta.Name, &meta_v1.DeleteOptions{})
}

func (c *Cleaner) updatePod(old, new interface{}) {
  oldPod := old.(*corev1.Pod)
  newPod := new.(*corev1.Pod)
  if oldPod.ResourceVersion == newPod.ResourceVersion {
    return
  }
  // If pod was ready but the current status is not ready then we are interested in it...
  //TODO: what happens when a Node loses connection, but comes back before Pod timeout?
  //      does the Pod changes state? do we accidentally cleanup?
  if !isPodRunning(newPod) && isPodRunning(oldPod) {
    c.enqueuePod(new)
  }
}

func isPodRunning(pod *corev1.Pod) bool {
  return pod.Status.Phase == corev1.PodRunning
}

func (c *Cleaner) enqueuePod(obj interface{}) {
  var key string
  var err error
  if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
    log.Println("WARNING: Could not schedule Pod for automatic cleanup because:" + err.Error())
    return
  }
  c.workqueue.Add(key)
}

func (c *Cleaner) delPod(obj interface{}) {
  var object meta_v1.Object
  var ok bool
  if object, ok = obj.(meta_v1.Object); !ok {
    tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
    if !ok {
      return
    }
    object, ok = tombstone.Obj.(meta_v1.Object)
    if !ok {
      return
    }
  }
  c.enqueuePod(obj)
}
