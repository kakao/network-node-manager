# IPVS Node Controller 

ipvs-node-controller is the kubernetes controller that solves External-IP (Load Balancer IP) issue with IPVS proxy mode. IPVS proxy mode has various problems, and one of them is that the External-IP assigned through the LoadBalancer type service with externalTrafficPolicy=Local option cannot access inside the cluster. More details on this issue can be found at [here](https://github.com/kubernetes/kubernetes/issues/75262). ipvs-node-controller solves this issue. ipvs-node-controller is based on [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder).

## Deploy

ipvs-node-controller now supports below CPU architectures.

* amd64
* arm64

Deploy ipvs-node-controllers through below command.

```
kubectl apply -f https://raw.githubusercontent.com/kakao/ipvs-node-controller/master/deploy/ipvs-node-controller.yml
```

## How it works?

ipvs-node-controller works on all worker nodes and adds the DNAT rules that converts destination IP of a packet from External-IP to Cluster-IP to iptables. ipvs-node-controller adds two DNAT rules for each LoadBalancer type service. One is added to the prerouting chain and the other is added to the output chain. The DNAT rule in the prerouting chain is for the pod that uses pod-only network namespace. On the other hand, The DNAT rule in the output chain is for the pod that uses host network namespace. All DNAT rules only target packets from pods on the host. Below is an example.

```
$ kubectl -n default get service 
NAME           TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                      AGE
lb-service-1   LoadBalancer   10.231.42.164   10.19.20.201   80:31751/TCP,443:30126/TCP   16d
lb-service-2   LoadBalancer   10.231.2.62     10.19.22.57    80:32352/TCP,443:31549/TCP   16d

$ iptables -nvL -t nat
...
Chain IPVS_OUTPUT (1 references)
 pkts bytes target     prot opt in     out     source               destination
    0     0 KUBE-MARK-MASQ  all  --  *      *       0.0.0.0/0            10.19.20.201         /* default/lb-service-1 */ ADDRTYPE match src-type LOCAL
    0     0 DNAT       all  --  *      *       0.0.0.0/0            10.19.20.201         /* default/lb-service-1 */ ADDRTYPE match src-type LOCAL to:10.231.42.164
    2   120 KUBE-MARK-MASQ  all  --  *      *       0.0.0.0/0            10.19.22.57          /* default/lb-service-2 */ ADDRTYPE match src-type LOCAL
    2   120 DNAT       all  --  *      *       0.0.0.0/0            10.19.22.57          /* default/lb-service-2 */ ADDRTYPE match src-type LOCAL to:10.231.2.62

Chain IPVS_PREROUTING (1 references)
 pkts bytes target     prot opt in     out     source               destination
    0     0 KUBE-MARK-MASQ  all  --  *      *       10.240.2.128/25      10.19.20.201         /* default/lb-service-1 */
    0     0 DNAT       all  --  *      *       10.240.2.128/25      10.19.20.201         /* default/lb-service-1 */ to:10.231.42.164
    1    60 KUBE-MARK-MASQ  all  --  *      *       10.240.2.128/25      10.19.22.57          /* default/lb-service-2 */
    1    60 DNAT       all  --  *      *       10.240.2.128/25      10.19.22.57          /* default/lb-service-2 */ to:10.231.2.62
...
```

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
