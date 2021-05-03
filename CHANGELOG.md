# Changelog

# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.8.0] - 2021-05-03

This release coincides with Cilium 1.10 and has support for new API additions
added in Cilium. Some of the new API features include support for agent and
debug events (#537), as well as prelimary support for the experimental Hubble
Recorder API (#530). Both of these features are currenlty only available via
the local unix domain socket. Other API features include filtering by TCP flags
(#461), IP version (#505) and node name (#412).

Hubble CLI 0.8 also comes with improvements to the CLI utility, such as reading
flows and filtering flows from stdin (#524), more flexible timestamp format
printing (#509), support for Apple silicon (#488), as well as miscellaneous
flag improvements and additions (#411, #420, #421, #443). It also contains a
new `hubble list` subcommand which, when targeting Hubble Relay, lists all
Hubble enabled nodes (#427).

*Breaking Changes*

In accordance with semver 0.x releases, this release contains a few
breaking changes to the Hubble command-line interface:

* The new default Hubble API endpoint (specified with `--server`) is now
  `localhost:4245` to ease usage with Hubble Relay. To connect to the local
  unix domain socket, use `--server unix:///var/run/cilium/hubble.sock` or set
  the `HUBBLE_SERVER` environment variable (default within a Cilium container)
  (#535)
* The new default output format is now always `compact` regardless of being in
  follow-mode or not. To obtain the old table output in the `hubble observe`
  command, use `--output=table` (#536)
* The source of reply packets is now printed on the left side in the compact
  output format. Such flows are indicated with a `<-` arrow instead of `->`.
  Flows with an unknown direction now use the `<>` arrow in the compact output
  (#533).
* The hidden `hubble peers watch` command has been renamed to `hubble watch
  peers` (#542)

**Major Changes:**
* Add basic support for agent events (#442, @tklauser)
* Add subcommands for agent and debug events (#537, @tklauser)
* cmd/observe: support for filtering events based on tcp-flags (#461, @nyrahul)
* cmd: add node list subcommand to list hubble nodes with status (#427, @rolinh)

**Minor Changes:**
* change default address to localhost:4245 (#535, @rolinh)
* cmd/config: add shell completion support for keys for get|set|reset (#420, @rolinh)
* cmd/observe: add a new flag to allow specifying different time formats for timestamps (#509, @rolinh)
* cmd/observe: add all flags (#411, @rolinh)
* cmd/observe: Add node name filter (#412, @twpayne)
* cmd/observe: add shell completion support for various flags (#421, @rolinh)
* cmd/observe: add support for IP version filters (#505, @rolinh)
* cmd/observe: mark deprecated output flags as deprecated (#506, @rolinh)
* cmd/observer: add support for agent event sub-type filters (#465, @tklauser)
* cmd: Add record subcommand (#530, @gandro)
* cmd: improve command usage message by grouping related flags (#443, @rolinh)
* compact: Always print original source on the left (#533, @michi-covalent)
* make: build release binaries for darwin/arm64 (aka Apple silicon) (#488, @rolinh)
* printer: Add support for debug events (#473, @gandro)
* RFC: cmd/observe: set default output format to "compact" (#536, @rolinh)
* Support reading flows from stdin (#524, @michi-covalent)

**Bugfixes:**
* cmd: fix environment variable names for options with dashes (#407, @rolinh)

**Misc Changes:**
* Agent event follow-up fixes for #442 (#454, @tklauser)
* all: avoid using the deprecated io/ioutil package (#489, @rolinh)
* Automate release creation and artifacts publishing (#490, @rolinh)
* build(deps): bump actions/setup-go from v1 to v2.1.3 (#476, @dependabot[bot])
* build(deps): bump github.com/sirupsen/logrus from 1.7.0 to 1.8.1 (#525, @dependabot[bot])
* build(deps): bump github.com/spf13/cobra from 1.1.2 to 1.1.3 (#486, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.36.0 to 1.36.1 (#522, @dependabot[bot])
* build(deps): bump google.golang.org/protobuf from 1.25.0 to 1.26.0 (#518, @dependabot[bot])
* build: ensure that binaries are always statically built (#397, @rolinh)
* Bump alpine base image to 3.13 (#472, @tklauser)
* Bump github.com/cilium/cilium to pull in reworked agent/debug event API (#532, @tklauser)
* ci: Add CodeQL analysis (#475, @twpayne)
* ci: Add dependabot configuration (#474, @twpayne)
* ci: do not upload artifacts (#485, @rolinh)
* ci: fix dependabot kind/enhancement label (#477, @kaworu)
* cmd/node: fix completion of output flag (#466, @rolinh)
* cmd/node: Refactor & Test output methods (#496, @simar7)
* cmd/observe: don't list agent/debug events and recorder captures in event type filter (#534, @tklauser)
* cmd/observe: print filters in debug mode (#502, @rolinh)
* cmd/observe: Print the entire request in debug mode (#515, @michi-covalent)
* cmd/observe: use signal.NotifyContext to cancel context on SIGINT (#539, @rolinh)
* cmd/peer: Refactor and test processing of response (#499, @simar7)
* cmd: change "node list" command for "list node" (#541, @rolinh)
* cmd: change 'peers watch' command to 'watch peers' (#542, @rolinh)
* cmd: use config key constants instead of hardcoded strings (#471, @rolinh)
* completion: remove the copyright header (#444, @kaworu)
* doc: #hubble-devel on Slack is now #sig-hubble (#495, @rolinh)
* doc: fix broken links (#406, @rolinh)
* Dockerfile: use alpine 3.12 (#540, @aanm)
* docs: Point to stable documentation (#414, @joestringer)
* Ensure build with Cilium master (#463, @gandro)
* Fix brokenlink on README.md (#500, @kaitoii11)
* make: set missing IMAGE_TAG variable (#432, @rolinh)
* Makefile: Add support for DOCKER_FLAGS environment variable (#456, @jrajahalme)
* Move version into VERSION file (#434, @glibsm)
* readme: bump versions in releases table (#400, @rolinh)
* readme: update releases table, mark Hubble Relay as stable (#404, @rolinh)
* release: fix `release` binary usage instruction (#396, @rolinh)
* Revert "ci: fix dependabot kind/enhancement label" (#493, @kaworu)
* set version to 0.8.0-dev (#393, @rolinh)
* stable.txt: Bump to v0.7.0 (#405, @gandro)
* Switch protobuf module to google.golang.org/protobuf (#452, @tklauser)
* update CHANGELOG for releases v0.6.1 and v0.7.0 (#398, @rolinh)
* Update Go to 1.15.4 (#416, @rolinh)
* Update Go to 1.15.5 (#423, @tklauser)
* Update Go to 1.15.6 (#446, @tklauser)
* Update Go to 1.15.7 (#467, @tklauser)
* Update Go to 1.15.8 (#478, @tklauser)
* Update Go to 1.16.1 (#507, @tklauser)
* Update Go to 1.16.2 (#510, @rolinh)
* Update Go to 1.16.3 (#526, @tklauser)
* update Go to v1.16.0 (#487, @rolinh)
* update readme and stable.txt for v0.7.1 (#410, @rolinh)
* update release instructions (#399, @rolinh)
* Update RELEASE.md with `-dev` change (#520, @rolinh)
* vendor: bump Cilium and grpc (#538, @rolinh)
* vendor: bump github.com/cilium/cilium (#482, @rolinh)
* vendor: bump github.com/cilium/cilium (#528, @rolinh)
* vendor: bump github.com/google/go-cmp from 0.5.4 to 0.5.5 (#504, @rolinh)
* vendor: bump google.golang.org to v1.33.2 (#437, @tklauser)
* vendor: bump google.golang.org/grpc to v1.34.0 (#457, @tklauser)
* vendor: bump google.golang.org/grpc to v1.35.0 (#464, @tklauser)
* vendor: bump google.golang.org/grpc to v1.36.0 (#498, @rolinh)
* vendor: Bump gopkg.in/yaml.v2 to v2.4.0 (#441, @twpayne)
* vendor: bump honnef.co/go/tools from v0.1.1 to v0.1.2 (#494, @rolinh)
* vendor: bump honnef.co/go/tools from v0.1.2 to v0.1.3 (#513, @rolinh)
* vendor: bump honnef.co/go/tools to v0.1.1 (#484, @rolinh)

## [0.7.1] - 2020-10-22

**Bugfixes:**
* cmd: fix environment variable names for options with dashes (#408, @Rolinh)

**Misc Changes:**
* build: ensure that binaries are always statically built (#402, @Rolinh)

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
