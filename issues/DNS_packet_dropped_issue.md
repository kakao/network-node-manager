# DNS packet dropped issue

There is an issue in which some DNS packets are dropped due to the race condition of linux kernel conntrack. Because of this issue, a phenomenon in which Record Resolve of Service or Pod within Kubernetes Cluster often fails occurs.

## How to solve it

network-node-manager adds the not tracking rules for DNS packet to avoid the race condition of linux kernel conntrack. Below are rules set by network-node-manager. 

```
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
   97  7321 CT         udp  --  *      *       0.0.0.0/0            0.0.0.0/0            udp dpt:53 CT notrack
   76 10560 CT         udp  --  *      *       0.0.0.0/0            0.0.0.0/0            udp spt:53 CT notrack

Chain NMANAGER_NOT_DNS_PREROUTING (1 references)
 pkts bytes target     prot opt in     out     source               destination
   76  5744 CT         udp  --  *      *       0.0.0.0/0            0.0.0.0/0            udp dpt:53 CT notrack
   97 14596 CT         udp  --  *      *       0.0.0.0/0            0.0.0.0/0            udp spt:53 CT notrack
```

## Reference

* https://www.weave.works/blog/racy-conntrack-and-dns-lookup-timeouts
* https://github.com/weaveworks/weave/issues/3287
* https://github.com/kubernetes/kubernetes/issues/62628
* https://github.com/kubernetes/kubernetes/issues/56903
