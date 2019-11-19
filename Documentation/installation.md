# Installation of Hubble

## Requirements

 * [Cilium] Recommended: >= 1.7.0-rc1, Minimal: >= 1.6.3
 * [Helm]
 * [Kubernetes]

## Install Cilium

Install Cilium using the [Install instructions]. If you need help to
troubleshoot installation issues, ping us on the [Cilium Slack].

### Enable Datapath Aggregation

Hubble relies on on aggregation of events in the eBPF datapath of Cilium.
Please enable datapath aggregation by setting the value of
`monitor-aggregation` in the `cilium-config` ConfigMap to `medium` or higher:

    monitor-aggregation: medium

This is the default setting for new installs of Cilium 1.6 or later.

## Install Hubble

Generate the deployment files using [Helm] and deploy it:

    cd install/kubernetes
    helm template hubble \
        --namespace kube-system \
        --set metrics.enabled="{dns,drop,tcp,flow,port-distribution,icmp,http}" \
        > hubble.yaml

Configure Hubble (Optional):

 * [Configure metrics](metrics.md)

Deploy Hubble:

    kubectl apply -f hubble.yaml

## Optional: Enable L7 Visibility

 * [Enable DNS Visibility](dns_visibility.md)
 * [Enable HTTP Visibility](http_visibility.md)

## Next Steps

 * [Configure Metrics with Prometheus & Grafana](../tutorials/deploy-hubble-and-grafana/)
 * [Configure the service map UI](../tutorials/deploy-hubble-servicemap/)
 * [Explore Flow Queries](../tutorials/explore-flow-queries/)
 * [More Tutorials](../tutorials/README.md)

[Install instructions]: http://docs.cilium.io/en/stable/gettingstarted/#installation
[Cilium Slack]: https://slack.cilium.io/
[Helm]: https://helm.sh/
[Kubernetes]: https://kubernetes.io/
[Cilium]: https://github.com/cilium/cilium
