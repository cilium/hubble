# Quickstart

You can deploy Cilium and Hubble on minikube with 3 simple steps.

## Requirements

 * [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
 * [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

## Hubble in 3 Minutes or Less

### 1. Start minikube

    % minikube start --network-plugin=cni

### 2. Install Cilium

Download `quick-install.yaml` from Cilium repo:

    wget https://raw.githubusercontent.com/cilium/cilium/master/install/kubernetes/quick-install.yaml

and enable Hubble by adding:

    enable-hubble: "true"

to cilium-config configmap. Then, deploy Cilium daemonset:

    % kubectl apply -f ./quick-install.yaml

Once the Cilium pods starts, run `cilium status` command to verify that Hubble is enabled:

    % kubectl exec -n kube-system -t -c cilium-agent ds/cilium cilium status | grep Hubble
    Hubble:                 Ok              Current/Max Flows: 1084/4096 (26.46%), Flows/s: 3.61   Metrics: Disabled

### 3. Install Hubble

Deploy Hubble CLI daemonset:

    kubectl apply -f https://raw.githubusercontent.com/cilium/hubble/master/tutorials/explore-flow-queries/hubble-cli-minikube.yaml

Now you are ready to start observing flows:

    kubectl exec -n kube-system -t -c hubble ds/hubble -- hubble observe -f
