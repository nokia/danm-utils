module github.com/nokia/danm-utils

go 1.13

require (
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/containernetworking/plugins v0.8.5
	github.com/intel/multus-cni v0.0.0-20200316130803-079c853eba60 // indirect
	github.com/intel/sriov-cni v2.1.0+incompatible // indirect
	github.com/j-keck/arping v1.0.0 // indirect
	github.com/nokia/danm v0.0.0-20200829131925-94fce4202a96
	github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b // indirect
	github.com/vishvananda/netlink v1.1.1-0.20200221165523-c79a4b7b4066 // indirect
	k8s.io/api v0.19.0-beta.0
	k8s.io/apimachinery v0.19.0-beta.0
	k8s.io/client-go v0.18.3
	k8s.io/code-generator v0.18.3
	k8s.io/kubernetes v1.19.0-beta.0
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66
)

replace (
	k8s.io/api => k8s.io/api v0.18.3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.3
	k8s.io/apiserver => k8s.io/apiserver v0.18.3
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.3
	k8s.io/client-go => k8s.io/client-go v0.19.0-beta.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.3
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.3
	k8s.io/code-generator => k8s.io/code-generator v0.18.3
	k8s.io/component-base => k8s.io/component-base v0.18.3
	k8s.io/cri-api => k8s.io/cri-api v0.18.3
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.3
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.3
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.3
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.3
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.3
	k8s.io/kubectl => k8s.io/kubectl v0.18.3
	k8s.io/kubelet => k8s.io/kubelet v0.18.3
	k8s.io/kubernetes => k8s.io/kubernetes v1.19.0-beta.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.3
	k8s.io/metrics => k8s.io/metrics v0.18.3
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.3
)
