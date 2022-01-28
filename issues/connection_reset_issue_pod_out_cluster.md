# Connection Reset Issue between Pod and Out of Cluster

In an environment where the pod uses an **overlay network**, when a container transmits a packet to outside the cluster, the packet is SNAT'ed. After that, when receiving a response packet from outside the container, the response packet must be DNAT'ed and delivered to the pod. However, due to the conntrack bug of the linux kernel, the response packet is regarded as an **invalid packet**, so there is a problem that the respose packet is not DNAT'ed and transmiited to the host. The host considers the response packet sent to the host as an incorrectly transmitted packet and disconnects the TCP connection by sending a reset packet.

## How to solve it

network-node-manager adds DROP the rules for invalid packet to INPUT chain in filter table. Below are rules set by network-node-manager.

```
$ iptables -t filter -nvL
...
Chain INPUT (policy ACCEPT 3229 packets, 597K bytes)
 pkts bytes target     prot opt in     out     source               destination
 246K   56M NMANAGER_INPUT  all  --  *      *       0.0.0.0/0            0.0.0.0/0
...
Chain NMANAGER_INPUT (1 references)
 pkts bytes target     prot opt in     out     source               destination
 246K   56M NMANAGER_DROP_INVALID_INPUT  all  --  *      *       0.0.0.0/0            0.0.0.0/0

Chain NMANAGER_DROP_INVALID_INPUT (1 references)
 pkts bytes target     prot opt in     out     source               destination
    0     0 DROP       all  --  *      *       0.0.0.0/0            0.0.0.0/0            ctstate INVALID
```

## References

* https://github.com/moby/libnetwork/issues/1090 
