# Changelog

# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.7.0] - 2020-10-19

**Minor Changes:**
* Add config subcommand (#380, @Rolinh)
* Add reflect command (#378, @michi-covalent)
* cmd/observe: Add HTTP method and path filters (#371, @twpayne)
* cmd/peer: print tls.ServerName when available (#374, @Rolinh)
* cmd/status: Add flows per second to `hubble status` (#330, @gandro)
* cmd/status: print node availability information when available (#328, @Rolinh)
* cmd/status: report current/max flows on the same line (#346, @Rolinh)
* cmd: add support for fish and powershell completion (#316, @Rolinh)
* cmd: add support for TLS and mTLS (#372, @Rolinh)
* cmd: honor user configuration directory for the configuration file (#375, @Rolinh)
* cmd: remove globals, optimize grpc client conn creation, remove pprof (#369, @Rolinh)
* Dockerfile: Remove ENTRYPOINT (#355, @michi-covalent)
* printer: ommit node name from output (#358, @mdnix)
* Update Go to v1.15, drop support for darwin/386, add support for linux/[arm,arm64] (#343, @Rolinh)

**Bugfixes:**
* cmd/status: do not report flows ratio when max flows is zero (#345, @Rolinh)
* make: fix git hash variable assignments for old make versions (#290, @Rolinh)

**Misc Changes:**
* .gitattributes: hide go.sum and vendor/modules.txt in pull requests (#317, @Rolinh)
* actions: add go-mod check (#382, @Rolinh)
* Add staticcheck to `make check` (#344, @tklauser)
* Clarify wording in README (#341, @christarazi)
* cmd/config: only write provided key/value when using set subcommand (#385, @Rolinh)
* cmd: fix help message for the `-config` flag (#377, @Rolinh)
* cmd: update observe and status command description/formatting (#390, @Rolinh)
* defaults: avoid stutter in exported names (#383, @tklauser)
* docs: Add link to Cilium Development Guide (#376, @twpayne)
* Fixes SC2038 in check-fmt.sh (#360, @nebril)
* make: fix release build directory ownership (#321, @kAworu)
* make: vendor in ineffassign, staticcheck, and golint (#357, @kAworu)
* observe: Document default flow count output (#318, @joestringer)
* printer: avoid duplicate import (#342, @tklauser)
* printer: use fmt.Fprintln instead of fmt.Fprintf (#347, @tklauser)
* README: fix broken link to metrics documentation (#327, @Rolinh)
* Readme: remove old beta warning and make a components table (#322, @glibsm)
* README: Update links (#351, @pchaigno)
* Remove version from release artifact file names (#293, @michi-covalent)
* tutorials: Fix README.md (#340, @jrajahalme)
* Update Cilium dep and fix unit tests that subsequently broke (#335, @Rolinh)
* Update Go to 1.15.3 (#386, @tklauser)
* update Go version to v1.14.7 (#336, @Rolinh)
* update Go version to v1.15.2 (#365, @Rolinh)
* v0.7: vendor: bump cilium to v1.9.0-rc2 to track cilium v1.9 branch (#394, @Rolinh)
* vendor: bump cilium to master right before branching v1.9 (#392, @Rolinh)
* vendor: bump cobra to v1.1.1 (#391, @twpayne)
* vendor: bump dependencies (#389, @Rolinh)
* vendor: go mod tidy && go mod vendor && go mod verify (#381, @Rolinh)
* vendor: update cilium@latest, viper@v1.7.1 (#373, @Rolinh)

**Other Changes:**
* Add little helper actions (#326, @glibsm)
* Add RELEASE.md with release checklist (#281, @glibsm)
* Add stable.txt (#299, @michi-covalent)
* add v0.6.0 release notes to changelog and bump version to 0.7.0-dev (#275, @Rolinh)
* Build release artifacts inside a container (#295, @michi-covalent)
* docs: Re-add images linked in README (#309, @gandro)
* Fix v0.6 branch link in README (#306, @gandro)
* Generate release binaries (#285, @Rolinh)
* Prepare for Cilium 1.8 (#305, @gandro)
* printer: Add jsonpb output (#302, @michi-covalent)
* Remove contrib/scripts/release.sh (#297, @michi-covalent)
* Require Cilium 1.7.x (#283, @tgraf)
* Update Go to v1.14.6 (#320, @Rolinh)
* update Go version to v1.14.4 and alpine base image to v3.12 (#278, @Rolinh)
* update Go version to v1.14.5 (#319, @Rolinh)
* vendor: cilium@master (#313, @glibsm)

## [0.6.1] - 2020-06-12

**Bugfixes:**
* make: fix git hash variable assignments for old make versions (#291, @michi-covalent)

**Misc Changes:**
* update Go version to v1.14.4 and alpine base image to v3.12 (#280, @Rolinh)

**Other Changes:**
* Backport https://github.com/cilium/hubble/pull/285 (#286, @michi-covalent)
* Prepare v0.6.1 release (#289, @michi-covalent)
* Require Cilium 1.7.x (#287, @michi-covalent)

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
