# Network Node Manager

network-node-manager is a kubernetes controller that controls the network configuration of a node to resolve network issues of kubernetes. By simply deploying and configuring network-node-manager, you can solve kubernetes network issues that cannot be resolved by kubernetes or resolved by the higher kubernetes Version. Below is a list of kubernetes's issues to be resolved by network-node-manager. network-node-manager is based on [kubebuilder v2.3.1](https://github.com/kubernetes-sigs/kubebuilder).

* [Connection reset issue between pod and out of cluster](issues/connection_reset_issue_pod_out_cluster.md)
* [External-IP access issue with IPVS proxy mode](issues/external_IP_access_issue_IPVS_proxy_mode.md)

## Deploy

network-node-manager now supports below CPU architectures.

* amd64
* arm64

Deploy network-node-manager through below command according to kube-proxy mode. After deploying network-node-manager, the "POD_CIDR_IPv4" environment variable must be set. And if you use IPv6 service, you must also set the "POD_CIDR_IPv6" environment variable.

```
iptables proxy mode 
$ kubectl apply -f https://raw.githubusercontent.com/kakao/network-node-manager/master/deploy/network-node-manager_iptables.yml
$ kubectl -n kube-system set env daemonset/network-node-manager POD_CIDR_IPV4=[IPv4 POD CIDR]
$ kubectl -n kube-system set env daemonset/network-node-manager POD_CIDR_IPV6=[IPv6 POD CIDR]

IPVS proxy mode
$ kubectl apply -f https://raw.githubusercontent.com/kakao/network-node-manager/master/deploy/network-node-manager_ipvs.yml
$ kubectl -n kube-system set env daemonset/network-node-manager POD_CIDR_IPV4=[IPv4 POD CIDR]
$ kubectl -n kube-system set env daemonset/network-node-manager POD_CIDR_IPV6=[IPv6 POD CIDR]
```

Examples are below

```
Example 1
$ kubectl apply -f https://raw.githubusercontent.com/kakao/network-node-manager/master/deploy/network-node-manager_iptables.yml
$ kubectl -n kube-system set env daemonset/network-node-manager POD_CIDR_IPV4="10.244.0.0/16"

Example 2
$ kubectl apply -f https://raw.githubusercontent.com/kakao/network-node-manager/master/deploy/network-node-manager_ipvs.yml
$ kubectl -n kube-system set env daemonset/network-node-manager POD_CIDR_IPV4="192.167.0.0/16"
$ kubectl -n kube-system set env daemonset/network-node-manager POD_CIDR_IPV6="fdbb::0/64"
```

## Configuration

The following are configurations related to rules managed by network-node-manager to solve the network issue of kubernetes. Please check the configuration and its related Rule. Network-node-manager is configured through environment variable configuration. When the environment variable is changed, the rule is dynamically set as network-node-manager is redeployed.

### Enable Drop Invalid Packet Rule in INPUT chain

* Related issue : [Connection reset issue between pod and out of cluster](issues/connection_reset_issue_pod_out_cluster.md)
* Default : true
* iptables proxy mode manifest : true
* IPVS proxy mode manifest : true

```
On
$ kubectl -n kube-system set env daemonset/network-node-manager RULE_DROP_INVALID_INPUT_ENABLE=true

Off
$ kubectl -n kube-system set env daemonset/network-node-manager RULE_DROP_INVALID_INPUT_ENABLE=false
```
### Enable External-IP to Cluster-IP DNAT Rule

* Related issue : [External-IP access issue with IPVS proxy mode](issues/external_IP_access_issue_IPVS_proxy_mode.md)
* Default : false
* iptables proxy mode manifest : false
* IPVS proxy mode manifest : true

```
On
$ kubectl -n kube-system set env daemonset/network-node-manager RULE_EXTERNAL_CLUSTER_ENABLE=true

Off
$ kubectl -n kube-system set env daemonset/network-node-manager RULE_EXTERNAL_CLUSTER_ENABLE=false
```

## How it works?

![network-node-manager Architecture](img/network-node-manager_Architecture.PNG)

network-node-manager runs on all kubernetes cluster nodes in host network namespace with network privileges and manage the node network configuration. network-node-manager watches the kubernetes object through kubenetes API server like a general kubernetes controller and manage the node network configuration. Now network-node-manager only watches service object.

## License

This software is licensed under the [Apache 2 license](LICENSE), quoted below.

Copyright 2020 Kakao Corp. <http://www.kakaocorp.com>

Licensed under the Apache License, Version 2.0 (the "License"); you may not
use this project except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
License for the specific language governing permissions and limitations under
the License.
