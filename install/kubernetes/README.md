# Deployment of Hubble alongside Cilium

This `hubble.yaml` contains a Kubernetes DaemonSet which will deploy Hubble as
a pod in the `kube-system` namespace on each node. It assumes that each node
is running Cilium with it's monitor socket available under the host path
`/var/run/cilium/`. This holds true if Cilium is deployed via
its standard YAML manifests.

## Standard Deployment

To deploy the Hubble server DaemonSet, use the following
[Helm](https://helm.sh/) command:

    helm template hubble \
        --namespace kube-system \
        > hubble.yaml
    kubectl apply -f hubble.yaml

Set `--listen-client-urls` option to start the Hubble server on a TCP port
instead of unix domain socket. For example, to listen to port 8080 on localhost:

    --set listenClientUrls='{localhost:8080}'

## Metrics

1. Metrics can be enabled via the `metrics.enabled` helm option:

        helm template hubble \
            --namespace kube-system \
            --set metrics.enabled="{dns:query,drop,tcp,flow,port-distribution}" \
            > hubble.yaml
        kubectl apply -f hubble.yaml

2. Deploy Prometheus & Grafana

   The following deploys Prometheus and Grafana into the `cilium-monitoring`
   namespace:

       kubectl apply -f https://raw.githubusercontent.com/cilium/cilium/v1.7/examples/kubernetes/addons/prometheus/monitoring-example.yaml

   Import the dashboard (`install/kubernetes/grafana.json`) via *Create* ->
   *Import*

## Usage

To query Hubble with the CLI on the first node:

    kubectl exec -n kube-system -t -c hubble ds/hubble -- hubble observe --since 1s
