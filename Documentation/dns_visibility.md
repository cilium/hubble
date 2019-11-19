# Enabling DNS Visibility

In order for Hubble to gain DNS visibility. Cilium must be configured to route
DNS traffic via its own DNS proxy to extract DNS information.

## With CiliumNetworkPolicy

The following CiliumNetworkPolicy will redirect the traffic of all pods to
CoreDNS via Cilium's DNS proxy to provide visibility. In order to not block any
non-DNS traffic, the policy below also allows all other traffic. Adjust the
policy accordingly if you are using other network policies.

    apiVersion: cilium.io/v2
    kind: CiliumNetworkPolicy
    metadata:
      name: dns-visibility
    spec:
      endpointSelector:
        matchLabels: {}
      egress:
      - toEndpoints:
        - matchLabels:
            k8s:io.kubernetes.pod.namespace: kube-system
            k8s:k8s-app: kube-dns
        toPorts:
        - ports:
          - port: "53"
            protocol: ANY
          rules:
            dns:
            - matchPattern: '*'
      - toFQDNs:
        - matchPattern: '*'
      - toEntities:
        - cluster
        - world

All network policies are specific to a namespace. So the policy will have to be
imported into each namespace separately.

## With Pod Annotations

An alternative to using CiliumNetworkPolicy, pods can be annotated to enable
visibility:

    kubectl annotate pod foo io.cilium.proxy-visibility="<Egress/53/UDP/DNS>"

**Important:** Visibility via annotations is not compatible with network
policies. You cannot use annotations while also using network policies.
