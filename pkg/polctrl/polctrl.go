package polctrl

import (
  "errors"
  "io"
  "log"
  "os"
  "time"
  polclientset "github.com/nokia/danm-utils/crd/client/clientset/versioned"
  polinformers "github.com/nokia/danm-utils/crd/client/informers/externalversions"
  meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  apierrors "k8s.io/apimachinery/pkg/api/errors"
  kubeinformers "k8s.io/client-go/informers"
//  coreinformers "k8s.io/client-go/informers/core/v1"
  "k8s.io/client-go/rest"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/cache"
)

const(
  MaxRetryCount = 5
  RetryInterval = 100
)

type NetPolControl struct {
  PolicyController cache.SharedIndexInformer
  PodController cache.SharedIndexInformer
  StopChan *chan struct{}
}

func NewNetPolControl(cfg *rest.Config, stopChan  *chan struct{}) (*NetPolControl,error) {
  polControl := &NetPolControl{StopChan: stopChan}
  polClient, err := polclientset.NewForConfig(cfg)
  if err != nil {
    return nil, err
  }
  for i := 0; i < MaxRetryCount; i++ {
    log.Println("INFO: Trying to discover DanmNetworkPolicy API in the cluster...")
    _, err = polClient.NetpolV1().DanmNetworkPolicies("").List(meta_v1.ListOptions{})
    if err != nil {
      log.Println("INFO: DanmNetworkPolicy discovery query failed with error:" + err.Error())
      time.Sleep(RetryInterval * time.Millisecond)
    } else {
      log.Println("INFO: DanmNetworkPolicy API seems to be installed in the cluster!")
      polControl.createPolicyController(polClient)
      break
    }
  }
  if polControl.PolicyController == nil {
    return nil, errors.New("DanmNetworkPolicy API is not installed in the cluster, DANM Network Policy Controller cannot start!")
  }
  polControl.createPodController(cfg)
  return polControl, nil
}

func (netpolController *NetPolControl) Run() {
  go netpolController.PolicyController.Run(*netpolController.StopChan)
  go netpolController.PodController.Run(*netpolController.StopChan)
}

func (netpolController *NetPolControl) WatchErrorHandler(r *cache.Reflector, err error) {
	if apierrors.IsResourceExpired(err) || apierrors.IsGone(err) || err == io.EOF {
    log.Println("INFO: One of the API watchers closed gracefully, re-establishing connection")
    return
  }
  //The default K8s client retry mechanism expires after a certain amount of time, and just gives-up
  //It is better to shutdown the whole process now and freshly re-build the watchers, than risking becoming a permanent zombie
  *netpolController.StopChan <- struct{}{}
  //Give some time for gracefully terminating the connections
  time.Sleep(5*time.Second)
  log.Println("ERROR: One of the API watchers closed unexpectedly with error:" + err.Error() + " shutting down DANM Network Policy Controller!!")
  os.Exit(0)
}

//TODO: implement event handlers
func AddNetPol(netpol interface{}) {}
func UpdateNetPol(oldNetpol, newNetpol interface{}) {}
func DeleteNetPol(netpol interface{}) {}
func AddPod(pod interface{}) {}
func UpdatePod(oldPod, newPod interface{}) {}

func (netpolController *NetPolControl) createPolicyController(polclient polclientset.Interface) {
  netpolInformerFactory := polinformers.NewSharedInformerFactory(polclient, time.Second*30)
  polController := netpolInformerFactory.Netpol().V1().DanmNetworkPolicies().Informer()
  polController.AddEventHandler(cache.ResourceEventHandlerFuncs{
      AddFunc: AddNetPol,
      UpdateFunc: UpdateNetPol,
      DeleteFunc: DeleteNetPol,
  })
  polController.SetWatchErrorHandler(netpolController.WatchErrorHandler)
  netpolController.PolicyController = polController
}

func (netpolController *NetPolControl) createPodController(cfg *rest.Config) {
  kubeClient, _ := kubernetes.NewForConfig(cfg)
  kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
  podController := kubeInformerFactory.Core().V1().Pods().Informer()
  podController.AddEventHandler(cache.ResourceEventHandlerFuncs{
      AddFunc: AddPod,
      UpdateFunc: UpdatePod,
  })
  podController.SetWatchErrorHandler(netpolController.WatchErrorHandler)
  netpolController.PodController = podController
}