# Welcome to DANM Utils!
## Table of Contents
* [Why utils?](#why-utils)
* [DANM Cleaner](#danm-cleaner)
* [DANM Policer](#danm-policer)
* [Showalloc](#showalloc)

## Why utils?
The core DANM platform - [https://github.com/nokia/danm](https://github.com/nokia/danm) - demonstrates that a concise E2E networking suite really brings value to any Kubernetes cluster - even if you don't need multiple interfaces!
Scope creep is a real thing though! As it usually happens with projects after they have seen considerable production usage, more and more interesting use-cases rear up their -ugly? beautiful?- heads.
What if we could do this awesome thing too? And that? Who wouldn't want to have network policing for multiple interfaces, different provider integrations, autonomous IP management etc?

New features all sound very exciting for developers, but they also increase operation complexity at the same time. To effectively manage this complexity without compromising on the user experience of the existing DANM suite, we decided to implement all future DANM platform enhancements as independent, loosely coupled, and most importantly optional Controllers / Operators.

Danm-utils is the project which houses this catalogue of value-adding DANM platform components.

## DANM Cleaner
Network management, and the CNI ecosystem has some known design flaws DANM aims to address. One of the most important Day 0 feature was a unique IPAM module, capable of handling discontinous, cluster-wide allocations.
As projects following in our footstep quickly learned this comes with the extra burden of efficiently managing a synchronized external state store at scale. What most people haven't realised yet is that keeping this store up for a prolonged period of time, in an asynchrous and "only" IT-reliable environment such as Kubernetes takes a lot of hard work.

Kubernetes's simplistic network management view boils down to two operations: adding a Pod, and deleting a Pod. But what happens with an already existing Pod and its -node-independent- allocations when a node looses network connectivity?
Or maybe the whole Node goes down?
Or maybe the whole cluster goes down?
The CNI architecture doesn't have an answer to these questions, and we have seen time and time again that old allocations weren't getting properly cleaned-up because the orchestrator didn't have an API to convey such changes.

Now enter Cleaner! This administrative component was created to monitor the cluster, and recognize when one of the aforementioned scenarios happen.
Cleaner is capable of determining if an old DANM allocation is dangling, or is still attached to an existing instance.
When dangling allocations are observed Cleaner also makes sure to reconcile the observed cluster state with the expected one.
This is particularly important in an environment where static IP allocations are used, as not being able to re-instantiate a component relying on a static allocation can even lead to a perpetual outage, requiring manual intervention.

Not anymore with Cleaner's self-healing magic! If you rely on DANM's powerful IPAM to manage the IP addresses in your cluster, and especially if you also use static allocations we strongly suggest installing Cleaner in your long-running, production environment.

Fore more information on installation, usage, and features refer to Cleaner's own user guide: TODO

## DANM Policer
Securing the applications' network access is a mandatory key feature in any production-grade cloud infrastructure. K8s provides the NetworkPolicy API to abstract this feature, however its implementation is left to CNI backends.  
Even the handful of CNIs which do support this API implemented it in a tightly coupled way, dependent on their own special ways of creating network interfaces.  
But how can you secure your networks in environments where your Pods require multiple interfaces, provided by different CNIs? Do you really need to choose between security and performance?  
Not anymore with DANM Policer! The utility is built upon the unique feature set of the DANM CNI platform which enabled the creation of a universal micro-segmentation backend for Kubernetes.Policer promises to isolate any network interfaces of your Pods, created via any CNIs!

Fore more information on installation, usage, and features refer to Policer's own user guide: [Policer User Guide](https://github.com/nokia/danm-utils/blob/master/policer_user_guide.md)

## Showalloc
Showalloc is a handy diagnostic tool which can be used to decode and show the IP allocations of DanmNets, ClusterNetworks and TenantNetworks.
It works as a standalone CLI installed on your host, or executed from a Pod thanks to relying only on remote Kubernetes API access (and of course a Service Profile).

Parameters:
-   `kubeconfig`: to specify K8s apiserver credential
-   `dnet`,  `cnet`,  `tnet`: to specify DANM network
-   `n`: to specify namespace, by default it is  `default`
-   `6`: to switch to IPv6 mode, by default the IPv4 mode is active
-   `showCIDR`: to show the specified sub-CIDR only
-   `a`: to show "free" and "out of pool range" IPs as well, by default these are suppressed

Example run:
```
# showalloc --kubeconfig .kube/config -dnet sriov-a -n kube-system
DANM network Kind: DanmNet
       IP version: IPv4
             CIDR: 10.100.100.0/24
       Pool start: 10.100.100.1
       Pool   end: 10.100.100.254
--------------------------------------------------------------
   10.100.100.0 : reserved (network base address)
   10.100.100.1 : allocated (Pod: danmtest-869444956b-vxshh)
   10.100.100.2 : allocated (Pod: danmtest-869444956b-gk2wh)
   10.100.100.3 : allocated (Pod: danmtest-869444956b-rrfpr)
   10.100.100.4 : allocated (Pod: danmtest-869444956b-z92pl)
   10.100.100.5 : allocated (Pod: danmtest-869444956b-2j2vp)
 10.100.100.255 : reserved (broadcast address)
```
