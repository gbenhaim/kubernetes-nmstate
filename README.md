# kubernetes-nmstate
Host-networking for kubernetes driven by kubernetes and executed by nmstate.
## The "Why"
With hybrid clouds, host-networking setup is becoming even more challenging. 
Different payloads have different networking requirements, and not everything could be satisfied as overlays on top of the main interface of the node (e.g. SR-IOV, L2, other L2).
The [CNI](https://github.com/containernetworking/cni) standard enables different solutions for connecting networks on the node with pods. 
Some of them are [part of the standard](https://github.com/containernetworking/cni), and there are some that support: [OVS bridges](https://github.com/kubevirt/ovs-cni), [SR-IOV](https://github.com/hustcat/sriov-cni), and more...
However, in all of these cases, the node must have the networks setup before the pod is scheduled. Setting up the networks in a dynamic and heterogenous cluster, with dynamic networking requirements, is a challenge by itself - and this is what this project is addressing.
## The "How"
We use [nmstate](https://nmstate.github.io/) to perform state driven network configuration on each node, as well as to report back its current state. 
The system defines 2 CRDs ([Custom Resource Definitions](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)): ```NodeNetworkState``` and ```NodeNetConfPolicy``` as described [here](https://docs.google.com/document/d/1282BcYjYGIIxQKgMYi3nQodB4ML_gw9BSs5AXB7QUtg/).
In the project we provide a 2 processes (one each CRD handling) that could be invoked manually by an external system (e.g. [Machine Config Operator](https://github.com/openshift/machine-config-operator)) and or run as node daemons (or as a Kubernetes [daemon set](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)) listening and handling modifications for the CRDs.
Demo:
[![asciicast](https://asciinema.org/a/225838.svg)](https://asciinema.org/a/225838)
### State Handler
When it starts on the node, it reads the list of ```NodeNetworkState``` CRDs, if no CRD exist for the node it is executed on, it will create one, and fill it with the output of ```nmstatectl show``` as the current status in the CRD. 
If a ```NodeNetworkState``` CRD exists for the node, it will try to enforce the desired state from the CRD (using: ```nmstatectl set```), and then report back current state.
When running in "client" mode, it has nothing more to do. When running as a daemon, it will:
- Detects an update to the ```NodeNetworkState``` CRD which apply to the node it is running on, then, it will try to reenforce the desired state, and report back the current one. In case it detects deletion of the ```NodeNetworkState``` CRD, it will re-create it with current state only.
- In case that the enforcement partially or completely failed, the daemon will retry to enforce it (with exponential back-off) until it succeeded, or the desired state is modified again 
- Even if enforcement was successful, the daemon will periodically poll the current state of the node, and will report it if any modification happened. If such modification is causing the desired state to be different than the current one, it will try to reenforce it (as described above).
> Notes:
> - The ```NodeNetworkState`` CRD has an "un-managed" indicator, allowing an administrator to stop all enforcement and reporting for a node.
> - The desire state could be created based on ```NodeNetConfPolicy``` CRDs (see below), or just manually set by an external system.
### Policy Handler
Can run in distributed or centralized mode. In case of distributed (default mode), it will only handle the ```NodeNetworkState``` CRDs of node it is executed on. In case of centralized mode, there has to be only one instance of it that run at the same time.
#### Distributed Mode
Upon invocation, it reads the list of ```NodeNetworkState``` CRDs, as well as the list of ```NodeNetConfPolicy``` CRDs. It will find which ```NodeNetworkState``` CRD is for the node it is running on. It will also find all ```NodeNetConfPolicy``` CRDs that apply to that node (based on their [affinity and toleration](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity) objects). 
From the interface match logic in the applicable ```NodeNetConfPolicy``` CRDs, and the list of interfaces taken from the current state of the node (in the ```NodeNetworkState``` CRD), it will create aggregated desired state object, and update it into the relevant ```NodeNetworkState``` CRD.
When running in "client" mode, it has nothing more to do. When running as a daemon, it will:
- Detects updates to ```NodeNetConfPolicy``` CRDs applicable for current node, and update ```NodeNetworkState``` CRD if needed.
- Detect an update to the current state of ```NodeNetworkState``` CRD for the node, and see if its desired state needs to be modified.
#### Centralized mode
Very similar to the distributed mode, but in this case, the client od daemon must handle policies and states for all nodes in the system.
## Getting Started
To deploy kubernetes-nmstate on the cluster call:
```
kubectl --kubeconfig ~/.kube/config apply -f manifests/deployment.yaml
```
This would install the CRDs, the state and policy daemon sets, and RBACs. Then call:
```
kubectl --kubeconfig ~/.kube/config get ds state-controller -n nmstate-default
```
To make sure that all node state daemons are running. Once a daemon runs on each node, call:
```
kubectl --kubeconfig ~/.kube/config get nodenetworkstates -n nmstate-default -o yaml
```
To get the current state of all nodes that belong to the cluster.

__TODO__ the policy part.

Please also check out our [User Guide](https://github.com/nmstate/kubernetes-nmstate/tree/master/docs/user-guide.md)
## Contributing
Contributions are welcomed! Please see [here](https://github.com/nmstate/kubernetes-nmstate/tree/master/docs/developer-guide.md)
