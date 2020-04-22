# Deployment of Hubble CLI alongside Cilium

This Helm chart generates a Kubernetes DaemonSet definition which will deploy Hubble
on each node Cilium is running. See [hubble/values.yaml](hubble/values.yaml) for the
list of configuration parameters.

## Standard Deployment

To deploy the Hubble DaemonSet, use the following [Helm](https://helm.sh/) command:

    helm template hubble | kubectl apply -f -

## Usage

To query Hubble with the CLI on the first node:

    kubectl exec -n kube-system -t -c hubble ds/hubble -- hubble observe -f
