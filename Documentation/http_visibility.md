# Enabling HTTP Visibility

In order for Hubble to gain HTTP visibility. Cilium must be configured to route
HTTP traffic via a Cilium managed Envoy proxy to extract HTTP information.

## With CiliumNetworkPolicy

The following CiliumNetworkPolicy will redirect the traffic of all pods on port
80 to the HTTP proxy. In order to not block any non-HTTP traffic, the policy
also allows all other traffic. Adjust the policy accordingly if you are using
other network policies.

    apiVersion: cilium.io/v2
    kind: CiliumNetworkPolicy
    metadata:
      name: http-visibility
    spec:
      endpointSelector:
        matchLabels: {}
      ingress:
      - fromEntities:
        - all
        toPorts:
        - ports:
          - port: "80"
            protocol: TCP
          rules:
            http:
            - {}
      - fromEntities:
        - all

All network policies are specific to a namespace. So the policy will have to be
imported into each namespace separately.

## With Pod Annotations

An alternative to using CiliumNetworkPolicy, pods can be annotated to enable
visibility:

    kubectl annotate pod foo io.cilium.proxy-visibility="<Ingress/80/TCP/HTTP>"

**Important:** Visibility via annotations is not compatible with network
policies. You cannot use annotations while also using network policies.
