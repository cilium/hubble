# Quickstart - Hubble in 3 Minutes or Less

## 0. Requirements

 * [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/)
 * [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
 * [Helm v3](https://helm.sh/docs/intro/install/)

## 1. Start minikube

```
minikube start --network-plugin=cni
```

## 2. Deploy Cilium

Deploy Cilium with Hubble enabled:

```
wget https://github.com/cilium/cilium/archive/master.zip
unzip master.zip
cd cilium-master
helm install cilium ./install/kubernetes/cilium \
  --namespace kube-system \
  --set global.hubble.enabled=true \
  --set global.hubble.cli.enabled=true \
  --set global.hubble.cli.image.tag=latest
```

Make sure `cilium` and `hubble-cli` pods are in ready state before proceeding:

```
kubectl get ds -n kube-system cilium hubble-cli
```

The output should look something like this:
```
NAME         DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
cilium       1         1         1       1            1           <none>          3m59s
hubble-cli   1         1         1       1            1           <none>          3m59s
```

### 3. Using the `hubble` CLI

Now you are ready to start observing flows:

```
kubectl exec -n kube-system -t -c hubble-cli ds/hubble-cli -- hubble observe -f
```
