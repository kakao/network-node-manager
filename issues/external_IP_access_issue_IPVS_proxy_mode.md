# External-IP Access Issue with IPVS Proxy Mode

When using **IPVS kube-proxy mode**, External-IP assigned through the LoadBalancer type service with externalTrafficPolicy=Local option **cannot be accessed from inside the cluster**. Currently kubernetes is not preparing any patch to fix this issue.

## How to solve it

network-node-manager adds the DNAT rules for each LoadBalancer type service. One is added to the prerouting chain and the other is added to the output chain. The DNAT rule in the prerouting chain is for the pod that uses pod-only network namespace. On the other hand, The DNAT rule in the output chain is for the pod that uses host network namespace. All DNAT rules only target packets from pods on the host. Below are example rules set by network-node-manager.

```
$ kubectl -n default get service
NAME           TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                      AGE
lb-service-1   LoadBalancer   10.231.42.164   10.19.20.201   80:31751/TCP,443:30126/TCP   16d
lb-service-2   LoadBalancer   10.231.2.62     10.19.22.57    80:32352/TCP,443:31549/TCP   16d

$ iptables -t nat -nvL
...
Chain PREROUTING (policy ACCEPT 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination
  190 20813 NMANAGER_PREROUTING  all  --  *      *       0.0.0.0/0            0.0.0.0/0

Chain OUTPUT (policy ACCEPT 3 packets, 180 bytes)
 pkts bytes target     prot opt in     out     source               destination
  651 41106 NMANAGER_OUTPUT  all  --  *      *       0.0.0.0/0            0.0.0.0/0
...
Chain NMANAGER_OUTPUT (1 references)
 pkts bytes target     prot opt in     out     source               destination
  650 41046 NMANAGER_EX_CLUS_OUTPUT  all  --  *      *       0.0.0.0/0            0.0.0.0/0

Chain NMANAGER_PREROUTING (1 references)
 pkts bytes target     prot opt in     out     source               destination
  190 20813 NMANAGER_EX_CLUS_PREROUTING  all  --  *      *       0.0.0.0/0            0.0.0.0/0
...
Chain NMANAGER_EX_CLUS_OUTPUT (1 references)
 pkts bytes target     prot opt in     out     source               destination
    0     0 KUBE-MARK-MASQ  all  --  *      *       0.0.0.0/0            10.19.20.201         /* default/lb-service-1 */ ADDRTYPE match src-type LOCAL
    0     0 DNAT       all  --  *      *       0.0.0.0/0            10.19.20.201         /* default/lb-service-1 */ ADDRTYPE match src-type LOCAL to:10.231.42.164
    2   120 KUBE-MARK-MASQ  all  --  *      *       0.0.0.0/0            10.19.22.57          /* default/lb-service-2 */ ADDRTYPE match src-type LOCAL
    2   120 DNAT       all  --  *      *       0.0.0.0/0            10.19.22.57          /* default/lb-service-2 */ ADDRTYPE match src-type LOCAL to:10.231.2.62

Chain NMANAGER_EX_CLUS_PREROUTING (1 references)
 pkts bytes target     prot opt in     out     source               destination
    0     0 KUBE-MARK-MASQ  all  --  *      *       192.167.0.0/16       10.19.20.201         /* default/lb-service-1 */
    0     0 DNAT       all  --  *      *       192.167.0.0/16       10.19.20.201         /* default/lb-service-1 */ to:10.231.42.164
    1    60 KUBE-MARK-MASQ  all  --  *      *       192.167.0.0/16       10.19.22.57          /* default/lb-service-2 */
    1    60 DNAT       all  --  *      *       192.168.0.0/16       10.19.22.57          /* default/lb-service-2 */ to:10.231.2.62
...
```

## References

* https://github.com/kubernetes/kubernetes/issues/75262

