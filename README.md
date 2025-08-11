 <picture>
   <source media="(prefers-color-scheme: light)" srcset="https://cdn.jsdelivr.net/gh/cilium/hubble@main/Documentation/images/hubble_logo.png" width="350" alt="Hubble Logo">
   <img src="https://cdn.jsdelivr.net/gh/cilium/hubble@main/Documentation/images/hubble_logo-dark.png" width="350" alt="Hubble Logo">
</picture>

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

----

# Network, Service & Security Observability for Kubernetes

- [What is Hubble?](#what-is-hubble)
- [Getting Started](#getting-started)
- [Features](#features)
  - [Service Dependency Graph](#service-dependency-graph)
  - [Metrics & Monitoring](#metrics--monitoring)
  - [Flow Visibility](#flow-visibility)
- [Get in touch / Community](#community)
- [Authors](#authors)
# What is Hubble?

Hubble is a fully distributed networking and security observability platform
for cloud native workloads. It is built on top of [Cilium] and [eBPF] to enable
deep visibility into the communication and behavior of services as well as the
networking infrastructure in a completely transparent manner.

Hubble can answer questions such as:

**Service dependencies & communication map:**
 * What services are communicating with each other? How frequently? What does
   the service dependency graph look like?
 * What HTTP calls are being made? What Kafka topics does a service consume
   from or produce to?

**Operational monitoring & alerting:**
 * Is any network communication failing? Why is communication failing? Is it
   DNS? Is it an application or network problem? Is the communication broken on
   layer 4 (TCP) or layer 7 (HTTP)?
 * Which services have experienced a DNS resolution problems in the last 5
   minutes? Which services have experienced an interrupted TCP connection
   recently or have seen connections timing out? What is the rate of unanswered
   TCP SYN requests?

**Application monitoring:**
 * What is the rate of 5xx or 4xx HTTP response codes for a particular service
   or across all clusters?
 * What is the 95th and 99th percentile latency between HTTP requests and
   responses in my cluster? Which services are performing the worst? What is
   the latency between two services?

**Security observability:**
 * Which services had connections blocked due to network policy? What services
   have been accessed from outside the cluster? Which services have resolved a
   particular DNS name?

## Why Hubble?

The Linux kernel technology [eBPF] is enabling visibility into systems and
applications at a granularity and efficiency that was not possible before. It
does so in a completely transparent way, without requiring the application to
change or for the application to hide information. By building on top of
[Cilium], Hubble can leverage [eBPF] for visibility. By leveraging [eBPF], all
visibility is programmable and allows for a dynamic approach that minimizes
overhead while providing deep and detailed insight where required. Hubble has
been created and specifically designed to make best use of these new [eBPF]
powers.

## Releases

The Hubble CLI is backward compatible with all supported Cilium releases. For
this reason, only the latest Hubble CLI version is maintained.

| Version                                              | Release Date         | Maintained | Supported Cilium Version | Artifacts                                                               |
|------------------------------------------------------|----------------------|------------|--------------------------|-------------------------------------------------------------------------|
| [v1.18](https://github.com/cilium/hubble/tree/main)  | 2025-08-11 (v1.18.0) | Yes        | Cilium 1.18 and older    | [GitHub Release](https://github.com/cilium/hubble/releases/tag/v1.18.0) |

## Component Stability

Hubble project consists of several components (see Architecture section).

While the core Hubble components have been running in production in multiple
environments, new components continue to emerge as the project grows and
expands in scope.

Some components, due to their relatively young age, are still considered beta
and have to be used with caution in critical production workloads.

| Component      | Area      | State  |
|----------------|-----------|--------|
| Hubble CLI     | Core      | Stable |
| Hubble Server  | Core      | Stable |
| Hubble Metrics | Core      | Stable |
| Hubble Relay   | Multinode | Stable |
| Hubble UI      | UI        | Beta   |

## Architecture

![Hubble Architecture](Documentation/images/hubble_arch.png)

# Getting Started

* [Introduction to Cilium & Hubble](https://docs.cilium.io/en/stable/overview/intro/)
* [Networking and Security Observability with Hubble](https://docs.cilium.io/en/stable/gettingstarted/hubble/)

# Features

## Service Dependency Graph

Troubleshooting microservices application connectivity is a challenging task.
Simply looking at "kubectl get pods" does not indicate dependencies between
each service or external APIs or databases.

Hubble enables zero-effort automatic discovery of the service dependency graph
for Kubernetes Clusters at L3/L4 and even L7, allowing user-friendly
visualization and filtering of those dataflows as a Service Map.

See [Hubble Service Map Tutorial](https://docs.cilium.io/en/stable/gettingstarted/hubble/)
for more examples.

![Service Map](Documentation/images/servicemap.png)

## Metrics & Monitoring

The metrics and monitoring functionality provides an overview of the state of
systems and allow to recognize patterns indicating failure and other scenarios
that require action. The following is a short list of example metrics, for a
more detailed list of examples, see the
[Metrics Documentation](https://docs.cilium.io/en/stable/observability/metrics/).

### Networking Behavior

![Networking](Documentation/images/network_and_tcp.png)

### Network Policy Observation

![Network Policy](Documentation/images/network_policy_pod.png)

### HTTP Request/Response Rate & Latency

![HTTP](Documentation/images/http.png)

### DNS Request/Response Monitoring

![DNS](Documentation/images/dns.png)

## Flow Visibility

Flow visibility provides visibility into flow information on the network and
application protocol level. This enables visibility into individual TCP
connections, DNS queries, HTTP requests, Kafka communication, and much more.

### DNS Resolution

Identifying pods which have received DNS response indicating failure:

    hubble observe --since=1m -t l7 -o json \
       | jq 'select(.l7.dns.rcode==3) | .destination.namespace + "/" + .destination.pod_name' \
       | sort | uniq -c | sort -r
      42 "starwars/jar-jar-binks-6f5847c97c-qmggv"

*Successful query & response:*

    starwars/x-wing-bd86d75c5-njv8k            kube-system/coredns-5c98db65d4-twwdg      DNS Query deathstar.starwars.svc.cluster.local. A
    kube-system/coredns-5c98db65d4-twwdg       starwars/x-wing-bd86d75c5-njv8k           DNS Answer "10.110.126.213" TTL: 3 (Query deathstar.starwars.svc.cluster.local. A)

*Non-existent domain:*

    starwars/jar-jar-binks-789c4b695d-ltrzm    kube-system/coredns-5c98db65d4-f4m8n      DNS Query unknown-galaxy.svc.cluster.local. A
    starwars/jar-jar-binks-789c4b695d-ltrzm    kube-system/coredns-5c98db65d4-f4m8n      DNS Query unknown-galaxy.svc.cluster.local. AAAA
    kube-system/coredns-5c98db65d4-twwdg       starwars/jar-jar-binks-789c4b695d-ltrzm   DNS Answer RCode: Non-Existent Domain TTL: 4294967295 (Query unknown-galaxy.starwars.svc.cluster.local. A)
    kube-system/coredns-5c98db65d4-twwdg       starwars/jar-jar-binks-789c4b695d-ltrzm   DNS Answer RCode: Non-Existent Domain TTL: 4294967295 (Query unknown-galaxy.starwars.svc.cluster.local. AAAA)

### HTTP Protocol

*Successful request & response with latency information:*

    starwars/x-wing-bd86d75c5-njv8k:53410      starwars/deathstar-695d8f7ddc-lvj84:80    HTTP/1.1 GET http://deathstar/
    starwars/deathstar-695d8f7ddc-lvj84:80     starwars/x-wing-bd86d75c5-njv8k:53410     HTTP/1.1 200 1ms (GET http://deathstar/)

### TCP/UDP Packets

*Successful TCP connection:*

    starwars/x-wing-bd86d75c5-njv8k:53410      starwars/deathstar-695d8f7ddc-lvj84:80    TCP Flags: SYN
    deathstar.starwars.svc.cluster.local:80    starwars/x-wing-bd86d75c5-njv8k:53410     TCP Flags: SYN, ACK
    starwars/x-wing-bd86d75c5-njv8k:53410      starwars/deathstar-695d8f7ddc-lvj84:80    TCP Flags: ACK, FIN
    deathstar.starwars.svc.cluster.local:80    starwars/x-wing-bd86d75c5-njv8k:53410     TCP Flags: ACK, FIN

*Connection timeout:*

    starwars/r2d2-6694d57947-xwhtz:60948   deathstar.starwars.svc.cluster.local:8080     TCP Flags: SYN
    starwars/r2d2-6694d57947-xwhtz:60948   deathstar.starwars.svc.cluster.local:8080     TCP Flags: SYN
    starwars/r2d2-6694d57947-xwhtz:60948   deathstar.starwars.svc.cluster.local:8080     TCP Flags: SYN

### Network Policy Behavior

*Denied connection attempt:*

    starwars/enterprise-5775b56c4b-thtwl:37800   starwars/deathstar-695d8f7ddc-lvj84:80(http)   Policy denied (L3)   TCP Flags: SYN
    starwars/enterprise-5775b56c4b-thtwl:37800   starwars/deathstar-695d8f7ddc-lvj84:80(http)   Policy denied (L3)   TCP Flags: SYN
    starwars/enterprise-5775b56c4b-thtwl:37800   starwars/deathstar-695d8f7ddc-lvj84:80(http)   Policy denied (L3)   TCP Flags: SYN

### Specifying Raw Flow Filters

Hubble supports extensive set of filtering options that can be specified as a combination of
allowlist and denylist. Hubble applies these filters as follows:

    for each flow:
      if flow does not match any of the allowlist filters:
        continue
      if flow matches any of the denylist filters:
        continue
      send flow to client

You can pass these filters to `hubble observe` command as
[JSON-encoded](https://developers.google.com/protocol-buffers/docs/proto3#json)
[FlowFilters](https://github.com/cilium/cilium/blob/v1.10.5/api/v1/flow/flow.proto#L348). For
example, to observe flows that match the following conditions:

- Either the source or destination identity contains `k8s:io.kubernetes.pod.namespace=kube-system`
  or `reserved:host` label, AND
- Neither the source nor destination identity contains `k8s:k8s-app=kube-dns` label:

      hubble observe \
        --allowlist '{"source_label":["k8s:io.kubernetes.pod.namespace=kube-system","reserved:host"]}' \
        --allowlist '{"destination_label":["k8s:io.kubernetes.pod.namespace=kube-system","reserved:host"]}' \
        --denylist '{"source_label":["k8s:k8s-app=kube-dns"]}' \
        --denylist '{"destination_label":["k8s:k8s-app=kube-dns"]}'

Alternatively, you can also specify these flags as `HUBBLE_{ALLOWLIST,DENYLIST}` environment variables:

    cat > allowlist.txt <<EOF
    {"source_label":["k8s:io.kubernetes.pod.namespace=kube-system","reserved:host"]}
    {"destination_label":["k8s:io.kubernetes.pod.namespace=kube-system","reserved:host"]}
    EOF

    cat > denylist.txt <<EOF
    {"source_label":["k8s:k8s-app=kube-dns"]}
    {"destination_label":["k8s:k8s-app=kube-dns"]}
    EOF

    HUBBLE_ALLOWLIST=$(cat allowlist.txt)
    HUBBLE_DENYLIST=$(cat denylist.txt)
    export HUBBLE_ALLOWLIST
    export HUBBLE_DENYLIST

    hubble observe

Note that `--allowlist` and `--denylist` filters get included in the request **in addition to**
regular flow filters like `--label` or `--namespace`. Use `--print-raw-filters` flag to verify
the exact filters that the Hubble CLI generates. For example:

    % hubble observe --print-raw-filters \
        -t drop \
        -n kube-system \
        --not --label "k8s:k8s-app=kube-dns" \
        --allowlist '{"source_label":["k8s:k8s-app=my-app"]}'
    allowlist:
    - '{"source_pod":["kube-system/"],"event_type":[{"type":1}]}'
    - '{"destination_pod":["kube-system/"],"event_type":[{"type":1}]}'
    - '{"source_label":["k8s:k8s-app=my-app"]}'
    denylist:
    - '{"source_label":["k8s:k8s-app=kube-dns"]}'
    - '{"destination_label":["k8s:k8s-app=kube-dns"]}'

The output YAML can be saved to a file and passed to `hubble observe` command with `--config`
flag. For example:

    % hubble observe --print-raw-filters --allowlist '{"source_label":["k8s:k8s-app=my-app"]}' > filters.yaml
    % hubble observe --config ./filters.yaml

# Community

Join the [Cilium Slack #hubble channel](https://slack.cilium.io) to chat
with Cilium Hubble developers and other Cilium / Hubble users. This is a good
place to learn about Hubble and Cilium, ask questions, and share your
experiences.

Learn more about [Cilium].

# Authors

Hubble is an open source project licensed under the [Apache License]. Everybody
is welcome to contribute. The project is following the [Governance Rules] of
the [Cilium] project. See [CONTRIBUTING] for instructions on how to contribute
and details of the Code of Conduct.


[Cilium]: https://github.com/cilium/cilium
[eBPF]: https://ebpf.io
[Apache License]: https://www.apache.org/licenses/LICENSE-2.0
[Governance Rules]: https://docs.cilium.io/en/stable/contributing/development/contributing_guide/
[CONTRIBUTING]: CONTRIBUTING.md
