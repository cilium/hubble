# Changelog

# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.6.0] - 2020-05-29

**Bugfixes:**
* api: fix potential panic in endpoint's EqualsByID (#199, @Rolinh)

**Misc Changes:**
* cmd: add hidden 'peer' command (#248, @Rolinh)
* update Go version to v1.14.2 (#226, @Rolinh)
* update Go version to v1.14.3 (#258, @Rolinh)

**Other Changes:**
* actions: Trigger on release branches (#233, @michi-covalent)
* Add changelog (#203, @glibsm)
* add peer gRPC service (#212, @Rolinh)
* Add support for policy verdict events (#200, @gandro)
* adjust dockerfile and makefile for "serve" command removal (#263, @Rolinh)
* Adjust to moved PolicyMatchType location (#222, @tgraf)
* api: Small fixes to the protoc invocations in Makefile (#206, @gandro)
* Bring back HUBBLE_DEFAULT_SOCKET_PATH env var (#239, @gandro)
* cmd/observe: use flags.DurationVar instead of StringVar for timeout flag (#210, @Rolinh)
* cmd/serve: refactor, introduce Server struct and options (#208, @Rolinh)
* cmd: Export RootCmd (#237, @glibsm)
* cmd: Finish config move (#254, @glibsm)
* cmd: Make all sub-commands more prominent (#255, @glibsm)
* cmd: Make pprof optional (#269, @gandro)
* cmd: Move completion and profile code from root (#246, @glibsm)
* defaults: Introduce new defaults for embedded Hubble (#224, @gandro)
* doc: Add a Quickstart section to the documentation (#243, @michi-covalent)
* doc: Update DNS visibility policy (#259, @michi-covalent)
* docker: ensure the hubble binary is statically built (#272, @Rolinh)
* fix: add skipped quote in hubble-all-minikube.yaml (#225, @geakstr)
* helm: Update hubble cli options (#245, @michi-covalent)
* l7: Add "Error" verdict (#211, @michi-covalent)
* make: optimize binary size by omitting symbol table and debug info (#268, @Rolinh)
* observe: Disable port-translation by default (#236, @michi-covalent)
* observe: Remove --port-translation (#271, @michi-covalent)
* observe: Show all the event types by default (#241, @michi-covalent)
* OnBuildFilter (#209, @tgraf)
* printer: Add support for NodeStatusEvent (#260, @gandro)
* printer: Fall back on ethernet MAC addresses (#261, @gandro)
* printer: Use policy verdict match type formatter from Cilium (#205, @gandro)
* Rebase vendored github.com/cilium (#232, @tgraf)
* Remove all server-side code (#220, @tgraf)
* Remove logger package (#221, @tgraf)
* server: Introduce per-request context (#216, @gandro)
* server: Match time range before filters (#213, @tgraf)
* Set version to 0.6.0-dev (#202, @glibsm)
* vendor: Bump github.com/cilium/cilium (#223, @gandro)
* vendor: pick up latest cilium (#247, @Rolinh)
* vendor: update cilium and sync replace directives (#207, @Rolinh)

## [0.5.0] - 2020-03-23
