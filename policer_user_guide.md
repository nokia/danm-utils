# DANM Policer user guide
## Introduction
Policer is an independent, optional Operator / Controller consuming core DANM APIs​. Policer's job is to enforce network policies in a DANM equipped Kubernetes environment, providing universal micro-segmentation capabilities in a multi-network Kubernetes environment.

Policer's main attraction is that it works even in a heterogenous environment, and makes network policies universally supported all on its own! Based on the unique way of how the core DANM platform manages networking, Policer can recongize and isolate all network interfaces of Pods, regardless which CNI plugins provisioned them!
    
Policer is controlled  by a dynamic CRD based Kubernetes API -very similar to NetworkPolicy-, and isolates Pods by provisioning iptables rules directly into their network namespace.

## Installation
### Building Policer
As is customary for any Kubernetes Controllers, Policer is also deployed as a Pod.
Building its container image is as easy as executing the following command in the root of the danm-utils project:

    ./build_policer.sh
The script first compiles the Policer binary in a builder container, and then installs it into an Alpine based container together with all the required packages.

### Deploying Policer
After the resulting image is onboarded to the target environment, Policer is ready to be deployed.
This process has two steps: deployment of the API, and the Controller.
Policer is controlled by a CRD based dynamic API called DanmNetworkPolicy. To onboard this API to your cluster, execute the following command on your target environment:

    kubectl create -f integration/crd/DanmNetworkPolicy.yaml

Now you are ready to deploy Policer itself, which is again as easy as executing the following command from the project root:

    kubectl create -f integration/manifests/policer/policer.yaml
This command creates all the Kubernetes artifacts required by the Policer DaemonSet, including ServiceAccounts, ClusterRoles, and ClusterRoleBindings.
Policer is an infrastructure component, so in order to do its job it requires elevated privileges. In order to deploy Policer you either must have the proper RBAC privileges allowing you to create these artifacts, or you need to contact your system administrator to do it for you.
All the Policer manifests are already following the principle of least privilege, and only ask elevated rights which are really required to be able to function.

## Usage
### DanmNetworkPolicy API
Policer once deployed can provision isolation rules into Pods. Abstract isolation rules can be provided via the DanmNetwokPolicy API.

DNP API is a literal  copy of the  NetworkPolicy API, and uses the exact  same  syntax for a familiar look&feel.
We decided to create a new API for two reasons:
-   Interoperability  with  CNIs  which  do  already consume  NetworkPolicy API
    
-   Multi-interface  support in rule  definitions for network selective  Ingress/Egress rules

As is the case with all networking related objects in Kubernetes, the default NetworkPolicy API lacks a network selector field. It simply works with the hard-coded assumption that isolation rules select "eth0" in the selected Pod.
This assumption however comes woefully short in a heterogenous, multi-network cluster. To be able to express network specific isolation rules, DanmNetworkPolicy API has one extra parameter compared to the upstream NetworkPolicy API called "NetworkSelector":
![DNP_API](https://github.com/nokia/danm-utils/blob/master/dnp_api.png)

### Interworking between the different selectors
#### Default behavior of the network selector in to/from rules
Regardless the addition of an extra selector option, the existing selectors work exactly as they do in upstream. All the interworking scenarios described in [Behavior of to and from selectors](https://kubernetes.io/docs/concepts/services-networking/network-policies/#behavior-of-to-and-from-selectors) document are supported, and work as defined by the Kubernetes standard: multiple policies and rules are additive, multiple selectors in the same rule are restrictive.
The only difference is that due to the addition of an optional network selector filter field the default interface selection mechanism changes from *"all interfaces but hey we anyway only have one, called eth0"* to *"literally all interfaces whatever their names"*.
This means that whenever a Pod is selected for whatever reasons and without an explicit network selector provided, DANM Policer will by default assume you want to whitelist **all** network interfaces of the Pod, rather than just the first.

To specifiy exactly what all network interfaces mean:
- any kernel network interfaces
- having a L3, IP address
- provisioned via any CNIs

will be considered by Policer.
#### Behavior of the network selector
As with all the other selectors, network selector can be used exclusively to define an isolation rule, or in combination with other selectors.

When the network selector is used alone, it selects all interfaces belonging to any Pods within the same namespace, which are connected to the referenced network. The network is identified via its name - DANM API type duplet.

When the network selector is used together with other selectors i.e. Pod selector, the filtering takes the logical AND of the subsets of the different selectors in accordance with the generic rules defined by the upstream Kubernetes standard. For example if Pod and network selectors are both defined, only interfaces selected by the network selector in the Pods selected by the Pod selector are whitelisted.

### Applying policies
#### Using network namespace iptables
Once Policer reached the decision that a Pod needs to be isolated, it provisions the isolation rules explained in the previous chapter.
Policer uses standard iptables / ip6tables utility to provision the isolation rules. As DANM knows the network namespace identifier of every Pod in the cluster, it can directly provision iptables rules into the network namespace of every Pod.
This is extremely advantageous for two reasons:
 - kernel interfaces provisioned by any CNI can be isolated this way, even those which do not have any "legs" in the host network namespace such as IPVLAN, or SR-IOV
 - as any given iptables instance in any given netns will only hold a small subset of the cluster's isolation rules, we will never hit the known performance and scalability bottlenecks of iptables
#### Iptables management
##### Policer created chains
Policer always creates its own chains for its own isolation rules for efficient rule management.
These chains are the following:
- "DANM_EGRESS_V4 in iptables to store "to" rules with IPv4 addresses
- "DANM_INGRESS_V4 in iptables to store "from" rules with IPv4 addresses
- "DANM_EGRESS_V6 in ip6tables to store "to" rules with IPv6 addresses
- "DANM_INGRESS_V6 in ip6tables to store "from" rules with IPv6 addresses
##### Default rules
Policer does not provision any rule for any Pod unless it is explicitly selected by a network policy.
When it is selected however, a set of default rules are added to it in addition to the user defined rules. These rules are required to ensure that the act of isolation does not unnecessarily hinder the normal communication flows of the Pod.
First of all, for every Policer created chains the following jump rules are added to the appropriate default INPUT/OUTPUT chains:

    Chain INPUT (policy ACCEPT 0 packets, 0 bytes)  
    pkts bytes target prot opt in out source destination  
    0 0 DANM_INGRESS_V4 all -- * * 0.0.0.0/0 0.0.0.0/0
    ...
    0 0 REJECT all -- * * 0.0.0.0/0 0.0.0.0/0 reject-with icmp-port-unreachable

 These rules ensure that only the packets explicitly whitelisted by Policer are allowed, and everything else is REJECTed, thus implementing default isolation for selected Pods. We use REJECT instead of DROP to protect against port scan type attacks fishing for TCP timeout events.
 Besides securing the INPUT and the OUTPUT chains, Policer also rejects all packets in the FORWARD chain.

In addition to rules providing default isolation, Policer also provisiong the following "quality of life" rules into the INPUT chains:

    0 0 ACCEPT all -- lo * 0.0.0.0/0 0.0.0.0/0  
    0 0 ACCEPT all -- * * 0.0.0.0/0 0.0.0.0/0 ctstate RELATED,ESTABLISHED
and these rules to the OUTPUT chains:

    0 0 ACCEPT all -- * lo 0.0.0.0/0 0.0.0.0/0  
    0 0 ACCEPT tcp -- * * 0.0.0.0/0 0.0.0.0/0 tcp dpt:53 ctstate NEW,ESTABLISHED  
    0 0 ACCEPT udp -- * * 0.0.0.0/0 0.0.0.0/0 udp dpt:53 ctstate NEW,ESTABLISHED

These rules ensure that basic functionalities are not accidentally broken even when no rules were explicitly defined to allow:
- Localhost Ingress and Engress communication
- Outgoing DNS client traffic (i.e. Services name resolution)
- Packets related to established ingress/egress dialogues with communication partners only allowed in one direction

Policer does not add any other rules to any default chains in the NAT table, or into any other table.
##### Dynamic rules
Apart from the few default rules Policer only adds rules to its own chains.
When an event is triggered, Policer reads all required API objects, parses them, and comes up with a streamlined set of rules to be provisioned in accordance with the selector logic explained earlier.

Every rule becomes exactly one entry in exactly one of the aforementioned chains. For every selected interface of every selected Pod Policer provisions an iptables rule explicitly allowing ingress, or egress communication to/from that IP by adding a rule with the IP set into -s / -d parameter.
If ports section is defined Policer creates extra rules for each mentioned ports using the selected interface's IP as the value for -s / -d parameter, plus the defined port(s) as -sport / -dport, and the defined protocol as -p.

In its current format Policer does not attempt to squash the rules into a more concise set. This optimization is something we might consider at a later stage, but even with this approach we don't expect to see major performance problems anyway due to the existence of the following pre-conditions:
- the rules are added to the Pod's netns, not to the host, therefore we don't expect to reach the iptables bottleneck thresholds even without squashing
- DANM already supports physically segregated networks, i.e. by using proper network selectors the number of entries needed to be added to each Pod can be concise in itself

Policer also doesn't try to validate whether adding a rule makes sense or not, it is dumb on purpose. Policer has no way to to know if L3 routing between two networks exists in the fabric or not, so even if two Pods are not connected to the same L2 segment they might still be able reach each other, making seemingly erroneous rules valid.

Policer fully supports provisioning rules for only V4, only V6, or dual-stack interfaces. When an interface of a Pod is selected as the target of a rule, Policer provisions one iptables rule for each IP found on the interface into the respective table. 
## Development
Policer is currently in an alpha phase. The base engine is implemented, and tested to work in practice. However, the engine isn't yet invoked during all lifecycle events when it is supposed to, and there are some restrictions as to which selector mechanism are currently supported.
You can check the current status of development under [Policer umbrella tracker](https://github.com/nokia/danm-utils/issues/7) 
If you like the idea of Policer and interested in pushing it forward, do not hesitate to join our Slack via https://join.slack.com/t/danmws/shared_invite/enQtNzEzMTQ4NDM2NTMxLTA3MDM4NGM0YTRjYzlhNGRiMDVlZWRlMjdlNTkwNTBjNWUyNjM0ZDQ3Y2E4YjE3NjVhNTE1MmEyYzkyMDRlNWU
