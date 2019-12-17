# Explore Flow Queries

## Basic Usage

    hubble observe [OPTIONS] [FILTERS]

# Kubernetes Helpers

The file [helpers.bash](../../install/kubernetes/helpers.bash) contains a set
of Bash functions which assist in using Hubble with Kubernetes. Source the file
in you Bash environment as follows:

``` bash
source install/kubernetes/helpers.bash
```

You can then run a Hubble CLI command across the entire cluster:

``` bash
hubble-cluster hubble observe --protocol dns --since=10m
```

Identify the Hubble pod name which runs on the same node as the pod specified:

``` bash
kubectl -n kube-system exec -ti $(hubble-pod default mypod-xxx) -- \
    hubble observe --since=5m --pod default/mypod-xxx -t drop
```

# Filters

Most filters support three variants to be specified `--field`, `--to-field`,
and `--from-field` to allow specifying whether the particular field should be
applied to the source or destination context, e.g. `--from-pod`, `--to-ip`,
`--ip`.

If multiple filters are provided, all filters must match for a flow to be
displayed.

## Negating filters

All filters can be negated by specifying `--not` before the filter

## Time Range

Follow flows as they are recorded

    hubble observe --follow

Using `--since` and `--until` it is possible to show the flows recoded in the
last 10 seconds, for example. Note that the possibility to look back in time is
limited by the size of the flow ring buffer.

    hubble observe --since=10s --until=5s

## Flow Type

Show only dropped flows

    hubble observe --type drop
    hubble observe -t drop

Show Layer 7 flows

    hubble observe --type l7

Show Cilium agent notifications and dropped flows

    hubble observe -t agent -t drop

## Kubernetes Pod Information

Show all flows originating from or destined to pod `default/my-app`.

    hubble observe --pod default/my-app
    hubble observe --from-pod default/app1 --to-pod default/app2

Show all flows originating from namespace `kube-system`:

    hubble observe --from-pod kube-system/
    hubble observe --from-namespace kube-system

Show all flows to a pod with a pod label `color=blue`

    hubble observe --to-label color=blue

## Verdict

Limit to forwarded or dropped flows (L3 & L7)

    hubble observe --verdict FORWARDED
    hubble observe --verdict DROPPED

## IP Addressing

Show flows from IP 10.1.1.1 to IP 192.168.1.1

    hubble observe --from-ip 10.1.1.1 --to-ip 192.168.1.1

Show flows to or from IP 30.1.1.1

    hubble observe --ip 30.1.1.1

## DNS

**Note:** This feature requires [DNS
Visibility](../Documentation/dns_visibility.md) to be enabled.

Show flows destined to an IP which was returned in response to a DNS request
querying a name matching `*.kuberntes.io` to be resolved.

    hubble observe --to-fqdn *.kubernetes.io

## HTTP

Show flows representing HTTP responses with status code `5xx`

    hubble observe --http-status 5+

# JSON Output

It is possible to output flows in JSON format using `-j`, one flow per row.

    hubble observe [FILTERS] -j

This in particular works well in combination with `jq`.

    hubble observe [FILTERS] -j | jq .
