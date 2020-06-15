# Installation of Hubble

## Requirements

 * [Cilium] >= 1.7.0, < 1.8.0 (for Hubble v0.5)
 * [Kubernetes]

## Install Cilium

Install Cilium using the [Install instructions]. To deploy Cilium 1.7 using quick-install.yaml:

    kubectl apply -f https://raw.githubusercontent.com/cilium/cilium/v1.7/install/kubernetes/quick-install.yaml

If you need help to troubleshoot installation issues, ping us on the
[Cilium Slack].

> Please note: If you are using Cilium 1.8 or newer, please check the Cilium
> documentation on how to enable the embedded Hubble server in Cilium >= 1.8.0.

### Enable Datapath Aggregation

Hubble relies on on aggregation of events in the eBPF datapath of Cilium.
Please enable datapath aggregation by setting the value of
`monitor-aggregation` in the `cilium-config` ConfigMap to `medium` or higher:

    monitor-aggregation: medium

This is the default setting for new installs of Cilium 1.6 or later.

## Install Hubble

Deploy Hubble using hubble-all-minikube.yaml:

    kubectl apply -f https://raw.githubusercontent.com/cilium/hubble/v0.5/tutorials/deploy-hubble-servicemap/hubble-all-minikube.yaml

### Optional: Configure Metrics

When you deploy Hubble with hubble-all-minikube.yaml, Hubble is configured with a default set of metric plugins. Follow
instructions in [this page](metrics.md) if you need to customize metric plugins.

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
