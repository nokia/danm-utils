package main

import (
  "flag"
  "fmt"
  "os"
  "time"
  kubeinformers "k8s.io/client-go/informers"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/kubernetes/scheme"
  "k8s.io/client-go/tools/leaderelection"
  "k8s.io/client-go/tools/record"
  "k8s.io/client-go/tools/leaderelection/resourcelock"
  "k8s.io/client-go/tools/clientcmd"
  utilruntime "k8s.io/apimachinery/pkg/util/runtime"
  v1core "k8s.io/client-go/kubernetes/typed/core/v1"
  corev1 "k8s.io/api/core/v1"
  danmclientset "github.com/nokia/danm/crd/client/clientset/versioned"
  danminformers "github.com/nokia/danm/crd/client/informers/externalversions"
  "github.com/nokia/danm-utils/pkg/cleaner"
)

var (
  kubeConf string
)

func main() {
  flag.StringVar(&kubeConf, "kubeconf", "", "Absolute path to a valid kubeconf file. Only required if Cleaner runs out-of-cluster.")
  flag.Parse()
  cfg, err := clientcmd.BuildConfigFromFlags("", kubeConf)
  if err != nil {
    log.Println("ERROR: cannot build cluster config for K8s REST client because:" + err.Error())
    return -1
  }
  kubeClient, err := kubernetes.NewForConfig(cfg)
  if err != nil {
    log.Println("ERROR: cannot build K8s REST client because:" + err.Error())
    return -1
  }
  danmClient, err := danmclientset.NewForConfig(cfg)
  if err != nil {
    log.Println("ERROR: cannot build DANM REST client because:" + err.Error())
    return -1
  }
  kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
  cleaner := cleaner.New(danmClient,
    kubeInformerFactory.Core().V1().Pods())
  cleanStuff := func(stopCh <-chan struct{}) {
    go kubeInformerFactory.Start(stopCh)
    if !cleaner.Initialize() {
      log.Println("ERROR: Cleaner timed-out synching its cache, retrying!")
      return -1
    }
    go cleaner.PeriodicCleanup(stopCh)
    if err = cleaner.Run(10, stopCh); err != nil {
      log.Println("ERROR: Cleaner failed with:" + err.Error())
      return -1
    }
  }
  rl, err := resourcelock.New(resourcelock.EndpointsResourceLock,
    "kube-system",
    "danm-cleaner",
    kubeClient.CoreV1(),
    resourcelock.ResourceLockConfig{
      Identity:      GetHostname(),
      EventRecorder: createRecorder(kubeClient, "danm-cleaner"),
    })
  if err != nil {
    log.Println("ERROR: Cannot create resource lock because:" + err.Error())
    return -1
  }
  leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
    Lock:          rl,
    LeaseDuration: 10 * time.Second,
    RenewDeadline: 5 * time.Second,
    RetryPeriod:   3 * time.Second,
    Callbacks: leaderelection.LeaderCallbacks{
      OnStartedLeading: cleanStuff,
      OnStoppedLeading: func() {
        utilruntime.HandleError(log.Println("WARNING: Cleaner cluster lost its leader"))
      },
    },
  })
  log.Println("WARNING: instance lost its lease, restarting!")
}

func GetHostname() string {
  ret, _ := os.Hostname()
  return ret
}

func createRecorder(kubeClient *kubernetes.Clientset, comp string) record.EventRecorder {
  eventBroadcaster := record.NewBroadcaster()
  eventBroadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: v1core.New(kubeClient.CoreV1().RESTClient()).Events("kube-system")})
  return eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: comp})
}