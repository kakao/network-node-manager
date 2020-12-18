# DNS packet dropped issue

There is an issue in which some DNS packets are dropped due to the race condition of linux kernel conntrack. Because of this issue, a phenomenon in which Record Resolve of Service or Pod within Kubernetes Cluster often fails occurs.

## How to solve it

This solution only works with **IPVS proxy** mode. network-node-manager adds the not tracking rules for DNS packet to avoid the race condition of linux kernel conntrack. The reason why notrack rule can be set for DNS packet is because Linux IPVS performs load balancing even for notrack packet. Since Linux iptables does not DNAT notrack packets, this solution cannot be applied when using iptables proxy mode. Unlike the [NodeLocal DNSCache](https://kubernetes.io/docs/tasks/administer-cluster/nodelocaldns/) solution, this solution has the advantage that it can be applied without restarting the kubelet or pod. Below are rules set by network-node-manager. 

```
$ kubectl -n kube-system get service kube-dns
NAME       TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)                  AGE
kube-dns   ClusterIP   10.96.0.10   <none>        53/UDP,53/TCP,9153/TCP   29m

$ kubectl get nodes worker -o jsonpath='{.spec.podCIDRs}'
["10.244.1.0/24"]

$ iptables -t raw -nvL
...
Chain PREROUTING (policy ACCEPT 9 packets, 643 bytes)
 pkts bytes target     prot opt in     out     source               destination
  215 18752 NMANAGER_PREROUTING  all      *      *       ::/0                 ::/0

Chain OUTPUT (policy ACCEPT 9 packets, 662 bytes)
 pkts bytes target     prot opt in     out     source               destination
  179 13607 NMANAGER_OUTPUT  all      *      *       ::/0                 ::/0
...
Chain NMANAGER_OUTPUT (1 references)
 pkts bytes target     prot opt in     out     source               destination
21109 2081K NMANAGER_NOT_DNS_OUTPUT  all  --  *      *       0.0.0.0/0            0.0.0.0/0

Chain NMANAGER_PREROUTING (1 references)
 pkts bytes target     prot opt in     out     source               destination
24587   26M NMANAGER_NOT_DNS_PREROUTING  all  --  *      *       0.0.0.0/0            0.0.0.0/0
...
Chain NMANAGER_NOT_DNS_OUTPUT (1 references)
 pkts bytes target     prot opt in     out     source               destination
    0     0 CT         udp  --  *      *       0.0.0.0/0            10.96.0.10           udp dpt:53 CT notrack
    0     0 CT         udp  --  *      *       0.0.0.0/0            10.244.1.0/24        udp dpt:53 CT notrack
    0     0 CT         udp  --  *      *       10.244.1.0/24        0.0.0.0/0            udp spt:53 CT notrack

Chain NMANAGER_NOT_DNS_PREROUTING (1 references)
 pkts bytes target     prot opt in     out     source               destination
    2   164 CT         udp  --  *      *       0.0.0.0/0            10.96.0.10           udp dpt:53 CT notrack
    0     0 CT         udp  --  *      *       0.0.0.0/0            10.244.1.0/24        udp dpt:53 CT notrack
    0     0 CT         udp  --  *      *       10.244.1.0/24        0.0.0.0/0            udp spt:53 CT notrack
```

## Reference

* https://www.weave.works/blog/racy-conntrack-and-dns-lookup-timeouts
* https://github.com/weaveworks/weave/issues/3287
* https://github.com/kubernetes/kubernetes/issues/62628
* https://github.com/kubernetes/kubernetes/issues/56903
