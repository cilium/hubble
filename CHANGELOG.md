# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.17.5] - 2025-06-23
[v1.17.3]: https://github.com/cilium/cilium/compare/v1.17.3...v1.17.5

**Misc Changes:**
* chore(deps): update actions/setup-go action to v5.5.0 (cilium/hubble#1680, @renovate[bot])
* chore(deps): update all github action dependencies (minor) (cilium/hubble#1683, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.17.4 (cilium/hubble#1682, @renovate[bot])
* chore(deps): update dependency helm/helm to v3.18.3 (cilium/hubble#1689, @renovate[bot])
* chore(deps): update dependency kubernetes-sigs/kind to v0.29.0 (cilium/hubble#1684, @renovate[bot])
* chore(deps): update library/golang docker tag to v1.24.4 (cilium/hubble#1687, @renovate[bot])
* chore(deps): update library/golang:1.24.3-alpine docker digest to b4f875e (cilium/hubble#1685, @renovate[bot])
* Update CONTRIBUTING.md (cilium/hubble#1681, @xmulligan)
* Update stable release to 1.17.3 (cilium/hubble#1677, @chancez)

## [v1.17.3] - 2025-04-30
[v1.17.3]: https://github.com/cilium/cilium/compare/v1.17.2...v1.17.3

**Misc Changes:**
* chore(deps): update all github action dependencies (patch) (cilium/hubble#1672, @renovate[bot])
* chore(deps): update library/golang docker tag to v1.23.8 (cilium/hubble#1671, @renovate[bot])
* chore(deps): update module golang.org/x/net to v0.38.0 [security] (cilium/hubble#1673, @renovate[bot])
* Update stable to 1.17.2 (cilium/hubble#1670, @chancez)

## [v1.17.2] - 2025-04-01
[v1.17.2]: https://github.com/cilium/cilium/compare/v1.17.1...v1.17.2

**Bugfixes:**
* hubble: escape terminal special characters from observe output (Backport PR cilium/cilium#37648, Upstream PR cilium/cilium#37401, @devodev)

**Misc Changes:**
* chore(deps): update actions/setup-go action to v5.4.0 (cilium/hubble#1667, @renovate[bot])
* chore(deps): update all github action dependencies (patch) (cilium/hubble#1666, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.17.1 (cilium/hubble#1661, @renovate[bot])
* chore(deps): update dependency helm/helm to v3.17.1 (cilium/hubble#1659, @renovate[bot])
* chore(deps): update dependency kubernetes-sigs/kind to v0.27.0 (cilium/hubble#1660, @renovate[bot])
* chore(deps): update library/golang docker tag to v1.23.7 (cilium/hubble#1664, @renovate[bot])
* chore(deps): update module golang.org/x/net to v0.36.0 [security] (cilium/hubble#1665, @renovate[bot])
* update stable release to 1.17.1 (cilium/hubble#1658, @rolinh)

## [v1.17.1] - 2025-02-12
[v1.17.1]: https://github.com/cilium/cilium/compare/v1.17.0...v1.17.1

**Minor Changes:**
* update Go to v1.23.6 and fix Renovate handling of Go (cilium/hubble#1650, @rolinh)

**Misc Changes:**
* migrate Renovate config (cilium/hubble#1652, @rolinh)
* migrate Renovate config take #2 (cilium/hubble#1654, @rolinh)
* Update stable release to 1.17.0 (cilium/hubble#1649, @rolinh)

## [v1.17.0] - 2025-02-07
[v1.17.0]: https://github.com/cilium/cilium/compare/v1.16.6...v1.17.0

**Minor Changes:**
* Add support for automatic port-forwarding in Hubble CLI Replace kubectl-based port-forwarding with native implementation in Cilium CLI (cilium/cilium#35483, @devodev)
* hubble: from and to cluster filters (cilium/cilium#33325, @kaworu)
* hubble: Stop building 32-bit binaries (cilium/cilium#35974, @michi-covalent)

**Bugfixes:**
* hubble: add printer for lost events (cilium/cilium#35208, @aanm)
* hubble: consistently use v as prefix for the Hubble version (cilium/cilium#35891, @rolinh)

**CI Changes:**
* Add Hubble CLI integration tests and skip running e2e/conformance on Hubble CLI only changes (cilium/cilium#33850, @chancez)

**Misc Changes:**
* .github: add cache to cilium-cli and hubble-cli build workflows (cilium/cilium#34847, @aanm)
* hubble: Add 'release' Make target (cilium/cilium#35561, @michi-covalent)
* hubble: Combine hubble and hubble-bin make targets (cilium/cilium#35256, @michi-covalent)
* hubble: remove outdated //go:build go1.18 tag (cilium/cilium#35174, @tklauser)
* hubble: Use hubble-bin target to generate release binaries (cilium/cilium#35127, @michi-covalent)
* make: add hubble cli to kind-image-fast-agent (cilium/cilium#35344, @kaworu)
* Refactor Hubble as a cell (cilium/cilium#35206, @kaworu)
* Remove deprecated call to DialContext in Hubble (cilium/cilium#34241, @davchos)
* Use Go standard library slices package more extensively (cilium/cilium#34796, @tklauser)
* chore(deps): update actions/setup-go action to v5.3.0 (cilium/hubble#1645, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.16.6 (cilium/hubble#1644, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.17.0 (cilium/hubble#1646, @renovate[bot])
* Update stable release to 1.16.6 (cilium/hubble#1643, @chancez)

## [v1.16.6] - 2025-01-22
[v1.16.6]: https://github.com/cilium/cilium/compare/v1.16.5...v1.16.6

**Misc Changes:**
* chore(deps): update all github action dependencies (minor) (cilium/hubble#1638, @renovate[bot])
* chore(deps): update all github action dependencies (patch) (cilium/hubble#1637, @renovate[bot])
* chore(deps): update dependency helm/helm to v3.17.0 (cilium/hubble#1641, @renovate[bot])
* chore(deps): update helm/kind-action action to v1.12.0 (cilium/hubble#1639, @renovate[bot])
* chore(deps): update module golang.org/x/net to v0.33.0 [security] (cilium/hubble#1636, @renovate[bot])
* Update readme/stable.txt to v1.16.5 (cilium/hubble#1635, @chancez)

## [v1.16.5] - 2024-12-18
[v1.16.5]: https://github.com/cilium/cilium/compare/v1.16.4...v1.16.5

**CI Changes:**
* Remove Dockerfile (cilium/hubble#1631, @michi-covalent)

**Misc Changes:**
* chore(deps): update actions/setup-go action to v5.2.0 (cilium/hubble#1633, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.16.4 (cilium/hubble#1624, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.21.0 (cilium/hubble#1627, @renovate[bot])
* chore(deps): update docker.io/library/alpine:3.21.0 docker digest to 21dc606 (cilium/hubble#1628, @renovate[bot])
* chore(deps): update docker/build-push-action action to v6.10.0 (cilium/hubble#1625, @renovate[bot])
* chore(deps): update golang to v1.23.4 (patch) (cilium/hubble#1626, @renovate[bot])
* release: Remove the step to post a Slack message (cilium/hubble#1622, @michi-covalent)
* Update stable release to 1.16.4 (cilium/hubble#1623, @michi-covalent)

## [v1.16.4] - 2024-11-20
[v1.16.4]: https://github.com/cilium/cilium/compare/v1.16.3...v1.16.4

**Misc Changes:**
* hubble: Add 'release' Make target (Backport PR cilium/cilium#35781, Upstream PR cilium/cilium#35561, @michi-covalent)
* chore(deps): update dependency helm/helm to v3.16.3 (cilium/hubble#1619, @renovate[bot])
* chore(deps): update dependency kubernetes-sigs/kind to v0.25.0 (cilium/hubble#1616, @renovate[bot])
* chore(deps): update docker.io/library/alpine:3.20.3 docker digest to 1e42bbe (cilium/hubble#1617, @renovate[bot])
* chore(deps): update golang (cilium/hubble#1618, @renovate[bot])
* chore(deps): update golang to v1.23.3 (patch) (cilium/hubble#1614, @renovate[bot])
* Update stable release to 1.16.3 (cilium/hubble#1611, @michi-covalent)

## [v1.16.3] - 2024-10-25
[v1.16.3]: https://github.com/cilium/cilium/compare/v1.16.2...v1.16.3

**Bugfixes:**
* hubble: add printer for lost events (Backport PR cilium/cilium#35274, Upstream PR cilium/cilium#35208, @aanm)

**Minor Changes:**
* .github: add cache to cilium-cli and hubble-cli build workflows (Backport PR cilium/cilium#35157, Upstream PR cilium/cilium#34847, @aanm)
* Makefile cleanups / improvements (cilium/hubble#1600, @michi-covalent)

**Misc Changes:**
* chore(deps): update actions/checkout action to v4.2.2 (cilium/hubble#1604, @renovate[bot])
* chore(deps): update actions/setup-go action to v5.1.0 (cilium/hubble#1605, @renovate[bot])
* chore(deps): update all github action dependencies (patch) (cilium/hubble#1602, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.16.3 (cilium/hubble#1603, @renovate[bot])
* Update stable release to 1.16.2 (cilium/hubble#1599, @michi-covalent)
* Update the release issue template (cilium/hubble#1597, @michi-covalent)

## [v1.16.2] - 2024-10-03
[v1.16.2]: https://github.com/cilium/cilium/compare/v1.16.1...v1.16.2

**Misc Changes:**
* chore(deps): update actions/checkout action to v4.2.0 (cilium/hubble#1590, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.16.2 (cilium/hubble#1589, @renovate[bot])
* chore(deps): update dependency helm/helm to v3.16.1 (cilium/hubble#1588, @renovate[bot])
* chore(deps): update dependency ubuntu to v24 (cilium/hubble#1591, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.20.3 (cilium/hubble#1587, @renovate[bot])
* chore(deps): update docker/build-push-action action to v6.8.0 (cilium/hubble#1592, @renovate[bot])
* chore(deps): update golang (cilium/hubble#1586, @renovate[bot])
* chore(deps): update golang to v1.23.2 (patch) (cilium/hubble#1593, @renovate[bot])
* Update stable release to 1.16.1 (cilium/hubble#1585, @glibsm)

## [v1.16.1] - 2024-09-11
[v1.16.1]: https://github.com/cilium/cilium/compare/v1.16.0...v1.16.1

**Misc Changes:**
* chore(deps): update actions/upload-artifact action to v4.4.0 (cilium/hubble#1582, @renovate[bot])
* chore(deps): update docker/build-push-action action to v6.6.1 (cilium/hubble#1576, @renovate[bot])
* chore(deps): update docker/setup-buildx-action action to v3.6.1 (cilium/hubble#1571, @renovate[bot])
* chore(deps): update golang to v1.23.1 (patch) (cilium/hubble#1583, @renovate[bot])

## [v1.16.0] - 2024-06-24
[v1.16.0]: https://github.com/cilium/cilium/compare/5aec7f58af0e57f93d5fa65f6e84a5e45609aac0...v1.16.0

**Major Changes:**
* Move cilium/hubble code to cilium/cilium repo (cilium/cilium#31893, @michi-covalent)

**Minor Changes:**
* hubble: node labels (cilium/cilium#32851, @kaworu)
* hubble: support drop\_reason\_desc in flow filter (cilium/cilium#32135, @chaunceyjiang)

**Misc Changes:**
* Add auto labeler for hubble-cli (cilium/cilium#32343, @aanm)
* hive: Rebase on cilium/hive (cilium/cilium#32020, @bimmlerd)
* hubble: Support --cel-expression filter in hubble observe (cilium/cilium#32147, @chancez)

## [v0.13.6] - 2024-06-18
[v0.13.6]: https://github.com/cilium/hubble/compare/v0.13.5...v0.13.6

**Minor Changes:**
* [v0.13] Bump golang to v1.21.11 (cilium/hubble#1517, @chancez)

**Misc Changes:**
* chore(deps): update all github action dependencies (v0.13) (minor) (cilium/hubble#1498, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (cilium/hubble#1497, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (cilium/hubble#1503, @renovate[bot])
* chore(deps): update docker/login-action action to v3.2.0 (v0.13) (cilium/hubble#1507, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.25.7 (v0.13) (cilium/hubble#1509, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.25.8 (v0.13) (cilium/hubble#1516, @renovate[bot])
* chore(deps): update golang stable (v0.13) (cilium/hubble#1496, @renovate[bot])
* Prepare for v0.13.5 development (cilium/hubble#1492, @gandro)

## [v0.13.5] - 2024-06-05
[v0.13.5]: https://github.com/cilium/hubble/compare/v0.13.4...v0.13.5

**Minor Changes:**
* [v0.13] Bump golang to v1.21.11 (cilium/hubble#1517, @chancez)

**Misc Changes:**
* chore(deps): update all github action dependencies (v0.13) (minor) (cilium/hubble#1498, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (cilium/hubble#1497, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (cilium/hubble#1503, @renovate[bot])
* chore(deps): update docker/login-action action to v3.2.0 (v0.13) (cilium/hubble#1507, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.25.7 (v0.13) (cilium/hubble#1509, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.25.8 (v0.13) (cilium/hubble#1516, @renovate[bot])
* chore(deps): update golang stable (v0.13) (cilium/hubble#1496, @renovate[bot])
* Prepare for v0.13.5 development (cilium/hubble#1492, @gandro)

## [v0.13.4] - 2024-05-13
[v0.13.4]: https://github.com/cilium/hubble/compare/v0.13.3...v0.13.4

**Misc Changes:**
* [v0.13] Bump Golang to v1.21.10 (#1491, @gandro)
* chore(deps): update actions/setup-go action to v5.0.1 (v0.13) (#1474, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (#1461, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (#1469, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (#1486, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.25.1 (v0.13) (#1462, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v5 (v0.13) (#1471, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v5.1.0 (v0.13) (#1475, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v5.3.0 (v0.13) (#1487, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v6 (v0.13) (#1488, @renovate[bot])
* chore(deps): update helm/kind-action action to v1.10.0 (v0.13) (#1470, @renovate[bot])
* Prepare for v0.13.4 and fix years in CHANGELOG (#1458, @glrf)
* vendor: Bump Cilium to v1.15.4 (#1489, @gandro)

## [v1.16.0-pre.2] - 2024-05-06
[v1.16.0-pre.2]: https://github.com/cilium/cilium/compare/v1.16.0-pre.1...v1.16.0-pre.2

**Major Changes:**
* Move cilium/hubble code to cilium/cilium repo (cilium/cilium#31893, @michi-covalent)

**Minor Changes:**
* hubble: support drop\_reason\_desc in flow filter (cilium/cilium#32135, @chaunceyjiang)

**Misc Changes:**
* hive: Rebase on cilium/hive (cilium/cilium#32020, @bimmlerd)
* hubble: Support --cel-expression filter in hubble observe (cilium/cilium#32147, @chancez)

## [v0.13.3] - 2024-04-18
[v0.13.3]: https://github.com/cilium/hubble/compare/v0.13.2...v0.13.3

**Misc Changes:**
* chore(deps): update all github action dependencies (v0.13) (minor) (#1422, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (#1421, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.15.3 (v0.13) (#1435, @renovate[bot])
* chore(deps): update dependency helm/helm to v3.14.4 (v0.13) (#1449, @renovate[bot])
* chore(deps): update docker/setup-buildx-action action to v3.3.0 (v0.13) (#1450, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.24.10 (v0.13) (#1443, @renovate[bot])
* chore(deps): update golang stable (v0.13) (#1428, @renovate[bot])
* Prepare for 0.13.3 (#1414, @chancez)
* v0.13: vendor: update cilium to v1.15.3 (#1433, @rolinh)

## [v0.13.2] - 2024-03-11
[v0.13.2]: https://github.com/cilium/hubble/compare/v0.13.1...v0.13.2

**Minor Changes:**
* Dockerfile: Update to Go 1.21.8 and Alpine 3.19.1 (#1412, @chancez)

**Misc Changes:**
* Prepare for v0.13.2 (#1408, @chancez)

## [v0.13.1] - 2024-03-08
[v0.13.1]: https://github.com/cilium/hubble/compare/v0.13.0...v0.13.1

**Misc Changes:**
* [v0.13] Prepare for v0.13.1 (#1351, @kaworu)
* chore(deps): update all github action dependencies (v0.13) (minor) (#1364, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (minor) (#1375, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (#1363, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (#1382, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (#1389, @renovate[bot])
* chore(deps): update all github action dependencies (v0.13) (patch) (#1394, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.15.0 (v0.13) (#1373, @renovate[bot])
* chore(deps): update dependency kubernetes-sigs/kind to v0.22.0 (v0.13) (#1390, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.23.2 (v0.13) (#1367, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.24.5 (v0.13) (#1396, @renovate[bot])
* chore(deps): update golang stable (v0.13) (#1362, @renovate[bot])
* chore(deps): update golang stable (v0.13) (#1372, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v4 (v0.13) (#1384, @renovate[bot])

## [v0.13.0] - 2024-01-15
[v0.13.0]: https://github.com/cilium/hubble/compare/v0.12.3...v0.13.0

**Minor Changes:**
* Add --cluster flag for filtering by cluster (#1309, @chancez)
* Add flags to filter for pods and services in any namespace (#1308, @glrf)
* Make `--namespace` flag more intuitive and allow both `--namespace` and `--pod` flag (#1279, @glrf)
* Add http header filter to hubble observe (#1277, @ChrsMark)
* Add HTTP URL filter (#1236, @glrf)
* Add request timeout flag for unary RPCs (#1290, @chancez)
* Display server provided flows rate (#1272, @glrf)

**Bugfixes:**
* Fix client certificate requests when no client certificate is specified (#1123, @chancez)
* Fix raw-filter flags by binding flags to viper only if command is run (#1202, @glrf)

**CI Changes:**
* ci: fix CodeQL workflow for v0.12, update deps in integration tests (#1131, @rolinh)
* ci: replace deprecated use of set-output (#1124, @rolinh)
* ci: run slowg analyzer (#1183, @rolinh)
* Renovate: Attempt to update Makefile golang with other Go deps (#1175, @chancez)
* renovate: Configure automerge only from trusted packages (#1278, @chancez)

**Misc Changes:**
* treewide: switch the logger to slog (#1108, @rolinh)
* build(deps): bump golang.org/x/net from 0.15.0 to 0.17.0 (#1251, @dependabot[bot])
* Add filter flag documentation and examples to help message (#1203, @glrf)
* Automerge Renovate minor/patch updates (#1263, @chancez)
* build: update docker.io/library/golang docker tag to alpine v3.19 (#1342, @kaworu)
* chore(deps): update actions/checkout action to v3.6.0 (main) (#1194, @renovate[bot])
* chore(deps): update actions/checkout action to v4 (main) (#1207, @renovate[bot])
* chore(deps): update actions/download-artifact action to v4.1.0 (main) (#1334, @renovate[bot])
* chore(deps): update actions/setup-go action to v5 (main) (#1313, @renovate[bot])
* chore(deps): update actions/upload-artifact action to v3.1.3 (main) (#1204, @renovate[bot])
* chore(deps): update all github action dependencies (main) (minor) (#1156, @renovate[bot])
* chore(deps): update all github action dependencies (main) (minor) (#1227, @renovate[bot])
* chore(deps): update all github action dependencies (main) (patch) (#1155, @renovate[bot])
* chore(deps): update all github action dependencies (main) (patch) (#1188, @renovate[bot])
* chore(deps): update all github action dependencies (main) (patch) (#1264, @renovate[bot])
* chore(deps): update all github action dependencies (main) (patch) (#1292, @renovate[bot])
* chore(deps): update all github action dependencies (main) (patch) (#1324, @renovate[bot])
* chore(deps): update all github action dependencies (master) (minor) (#1138, @renovate[bot])
* chore(deps): update all github action dependencies to v3 (main) (major) (#1219, @renovate[bot])
* chore(deps): update all github action dependencies to v4 (main) (major) (#1326, @renovate[bot])
* chore(deps): update dependency cilium/cilium to v1.14.2 (main) (#1224, @renovate[bot])
* chore(deps): update dependency go to v1.21.3 (main) (#1268, @renovate[bot])
* chore(deps): update dependency helm/helm to v3.12.3 (main) (#1184, @renovate[bot])
* chore(deps): update dependency helm/helm to v3.13.1 (main) (#1255, @renovate[bot])
* chore(deps): update dependency helm/helm to v3.13.2 (main) (#1286, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.18.3 (main) (#1179, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.18.4 (main) (#1231, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.18.5 (main) (#1303, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.19.0 (main) (#1318, @renovate[bot])
* chore(deps): update docker/build-push-action action to v4.2.0 (main) (#1208, @renovate[bot])
* chore(deps): update docker/build-push-action action to v4.2.1 (main) (#1214, @renovate[bot])
* chore(deps): update docker/build-push-action action to v5 (main) (#1220, @renovate[bot])
* chore(deps): update docker/build-push-action action to v5.1.0 (main) (#1297, @renovate[bot])
* chore(deps): update docker/setup-buildx-action action to v2.10.0 (main) (#1197, @renovate[bot])
* chore(deps): update docker/setup-buildx-action action to v2.9.1 (master) (#1126, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.20.4 (master) (#1136, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.0 (master) (#1148, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.3 (main) (#1181, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.5 (main) (#1196, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.7 (main) (#1217, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.8 (main) (#1225, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.9 (main) (#1232, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.2 (main) (#1240, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.3 (main) (#1260, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.4 (main) (#1270, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.5 (main) (#1275, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.8 (main) (#1302, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.9 (main) (#1311, @renovate[bot])
* chore(deps): update github/codeql-action action to v3 (main) (#1327, @renovate[bot])
* chore(deps): update github/codeql-action action to v3.23.0 (main) (#1341, @renovate[bot])
* chore(deps): update golang (main) (#1230, @renovate[bot])
* chore(deps): update golang (main) (#1281, @renovate[bot])
* chore(deps): update golang (main) (#1305, @renovate[bot])
* chore(deps): update golang (main) (#1307, @renovate[bot])
* chore(deps): update golang (main) (#1323, @renovate[bot])
* chore(deps): update golang (main) (#1333, @renovate[bot])
* chore(deps): update golang (main) (patch) (#1237, @renovate[bot])
* chore(deps): update golang to v1.20.6 (master) (patch) (#1127, @renovate[bot])
* chore(deps): update golang to v1.20.7 (main) (patch) (#1171, @renovate[bot])
* chore(deps): update golang to v1.21.0 (main) (minor) (#1177, @renovate[bot])
* chore(deps): update golang to v1.21.1 (main) (patch) (#1205, @renovate[bot])
* chore(deps): update golang to v1.21.3 (main) (patch) (#1256, @renovate[bot])
* chore(deps): update golang to v1.21.4 (main) (patch) (#1287, @renovate[bot])
* chore(deps): update golang to v1.21.5 (main) (patch) (#1310, @renovate[bot])
* chore(deps): update golang to v1.21.6 (main) (patch) (#1340, @renovate[bot])
* chore(deps): update golangci/golangci-lint docker tag to v1.54.1 (main) (#1187, @renovate[bot])
* chore(deps): update golangci/golangci-lint docker tag to v1.54.2 (main) (#1193, @renovate[bot])
* chore(deps): update golangci/golangci-lint docker tag to v1.55.0 (main) (#1267, @renovate[bot])
* chore(deps): update golangci/golangci-lint docker tag to v1.55.1 (main) (#1274, @renovate[bot])
* chore(deps): update golangci/golangci-lint docker tag to v1.55.2 (main) (#1284, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v3.7.0 (main) (#1189, @renovate[bot])
* chore(deps): update helm/kind-action action to v1.8.0 (master) (#1129, @renovate[bot])
* chore(deps): update library/golang docker tag to v1.20.6 (master) (#1128, @renovate[bot])
* ci: build image for arm64 (#1168, @rolinh)
* doc and config updates for the v0.12.0 release (#1121, @rolinh)
* doc: add server flag documentation to watch peers (#1229, @kaworu)
* doc: update doc readme and remove broken links (#1200, @rolinh)
* Don't set --last if --input-file is specified (#1153, @michi-covalent)
* enable new analyzers (protogetter, sloglint and testifylint) (#1295, @rolinh)
* fix(deps): update all go dependencies main (main) (#1154, @renovate[bot])
* fix(deps): update all go dependencies main (main) (minor) (#1206, @renovate[bot])
* fix(deps): update all go dependencies main (main) (minor) (#1288, @renovate[bot])
* fix(deps): update all go dependencies main (main) (minor) (#1304, @renovate[bot])
* fix(deps): update all go dependencies main (main) (patch) (#1319, @renovate[bot])
* fix(deps): update golang.org/x/exp digest to 613f0c0 (master) (#1125, @renovate[bot])
* fix(deps): update golang.org/x/exp digest to d63ba01 (main) (#1170, @renovate[bot])
* fix(deps): update golang.org/x/sys digest to 13b15b7 (main) (#1301, @renovate[bot])
* fix(deps): update module github.com/google/go-cmp to v0.6.0 (main) (#1258, @renovate[bot])
* fix(deps): update module github.com/spf13/viper to v1.17.0 (main) (#1241, @renovate[bot])
* fix(deps): update module github.com/spf13/viper to v1.18.0 (main) (#1312, @renovate[bot])
* fix(deps): update module golang.org/x/sys to v0.11.0 (main) (#1176, @renovate[bot])
* fix(deps): update module golang.org/x/sys to v0.13.0 (main) (#1238, @renovate[bot])
* fix(deps): update module golang.org/x/sys to v0.16.0 (main) (#1338, @renovate[bot])
* fix(deps): update module google.golang.org/grpc to v1.58.2 (main) (#1218, @renovate[bot])
* fix(deps): update module google.golang.org/grpc to v1.58.3 (main) (#1257, @renovate[bot])
* fix(deps): update module google.golang.org/grpc to v1.59.0 (main) (#1265, @renovate[bot])
* fix(deps): update module google.golang.org/grpc to v1.60.0 (main) (#1325, @renovate[bot])
* fix(deps): update module google.golang.org/protobuf to v1.32.0 (main) (#1336, @renovate[bot])
* Makefile: Support overriding build options via Makefile.override (#1285, @chancez)
* Release v0.12.2 follow up (#1254, @rolinh)
* Release v0.12.3 follow up (#1321, @lambdanis)
* renovate tweak (#1122, @kaworu)
* renovate: Add client-go to disabled list (#1167, @sayboras)
* renovate: Use allowedVersions to limit updates to stable branches (#1244, @chancez)
* switch the default git branch from master to main (#1147, @kaworu)
* treewide: use log/slog instead of golang.org/x/exp/slog (#1182, @rolinh)
* Update README and CHANGELOG to v0.12.1 (#1250, @gandro)
* vendor: bump Cilium to v1.14.0-rc.0 (#1120, @rolinh)
* vendor: Fix missing file in vendor (#1269, @chancez)
* vendor: update golang.org/x/sys to latest unreleased version (#1291, @rolinh)

## [v0.12.3] - 2023-12-08
[v0.12.3]: https://github.com/cilium/hubble/compare/v0.12.2...v0.12.3

**Misc Changes:**

* chore(deps): update actions/checkout action to v4.1.1 (v0.12) (#1266, @renovate[bot])
* chore(deps): update actions/setup-go action to v5 (v0.12) (#1316, @renovate[bot])
* chore(deps): update dependency go to v1.20.11 (v0.12) (#1289, @renovate[bot])
* chore(deps): update dependency go to v1.20.12 (v0.12) (#1314, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.18.5 (v0.12) (#1306, @renovate[bot])
* chore(deps): update docker/build-push-action action to v5.1.0 (v0.12) (#1298, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.2 (v0.12) (#1259, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.3 (v0.12) (#1261, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.4 (v0.12) (#1271, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.5 (v0.12) (#1276, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.8 (v0.12) (#1293, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.9 (v0.12) (#1315, @renovate[bot])
* chore(deps): update golang stable (v0.12) (#1282, @renovate[bot])
* deps: Update Go images to v1.20.12 (#1317, @lambdanis)

## [v0.12.2] - 2023-10-12
[v0.12.2]: https://github.com/cilium/hubble/compare/v0.12.1...v0.12.2

**Misc Changes:**

* vendor: update golang.org/x/net to v0.17.0 [security] (#1252, @rolinh)

## [v0.12.1] - 2023-10-11
[v0.12.1]: https://github.com/cilium/hubble/compare/v0.12.0...v0.12.1

**Misc Changes:**
* chore(deps): update actions/checkout action to v3.6.0 (v0.12) (#1195, @renovate[bot])
* chore(deps): update actions/checkout action to v4 (v0.12) (#1213, @renovate[bot])
* chore(deps): update actions/checkout action to v4.1.0 (v0.12) (#1228, @renovate[bot])
* chore(deps): update actions/setup-go action to v4.1.0 (v0.12) (#1180, @renovate[bot])
* chore(deps): update actions/upload-artifact action to v3.1.3 (v0.12) (#1209, @renovate[bot])
* chore(deps): update all github action dependencies (v0.12) (patch) (#1143, @renovate[bot])
* chore(deps): update all github action dependencies to v3 (v0.12) (major) (#1222, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.18.3 (v0.12) (#1178, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.18.4 (v0.12) (#1234, @renovate[bot])
* chore(deps): update docker/build-push-action action to v4.2.1 (v0.12) (#1211, @renovate[bot])
* chore(deps): update docker/build-push-action action to v5 (v0.12) (#1223, @renovate[bot])
* chore(deps): update docker/setup-buildx-action action to v2.10.0 (v0.12) (#1199, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.0 (v0.12) (#1149, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.3 (v0.12) (#1166, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.4 (v0.12) (#1190, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.5 (v0.12) (#1198, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.7 (v0.12) (#1221, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.8 (v0.12) (#1226, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.21.9 (v0.12) (#1235, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.22.1 (v0.12) (#1243, @renovate[bot])
* chore(deps): update golang stable (v0.12) (#1185, @renovate[bot])
* chore(deps): update golang stable (v0.12) (#1233, @renovate[bot])
* chore(deps): update golang stable to v1.20.10 (v0.12) (patch) (#1247, @renovate[bot])
* chore(deps): update golang stable to v1.20.6 (v0.12) (patch) (#1144, @renovate[bot])
* chore(deps): update golang stable to v1.20.7 (v0.12) (patch) (#1173, @renovate[bot])
* chore(deps): update golang stable to v1.20.8 (v0.12) (patch) (#1210, @renovate[bot])
* chore(deps): update golang to v1.21.1 (v0.12) (minor) (#1212, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v3.7.0 (v0.12) (#1191, @renovate[bot])
* chore(deps): update helm/kind-action action to v1.8.0 (v0.12) (#1146, @renovate[bot])
* chore(deps): update library/golang docker tag to v1.20.6 (v0.12) (#1145, @renovate[bot])
* Revert "chore(deps): update golang to v1.21.1" (#1245, @chancez)
* v0.12: ci: build image for arm64 (#1169, @rolinh)
* v0.12: vendor: update all deps to their latest patch release (#1248, @rolinh)

## [v0.12.0] - 2023-07-10
[v0.12.0]: https://github.com/cilium/hubble/compare/v0.11.6...v0.12.0

**Major Changes:**
* Add hubble list namespaces command (#1086, @chancez)
* Add support for supplying basic authentication credentials (#1002, @chancez)
* Replace stdin detection with --flows-file flag which supports reading flows from a file or stdin (#951, @chancez)

**Minor Changes:**
* Add experimental-field-mask flags (#1101, @AwesomePatrol)
* Correctly handle --first/--last when reading flows from a stdin (#958, @chancez)
* Hubble observe flows (#875, @chancez)
* Improve help message for --last option (#913, @PriyaSharma9)
* Log when connection is successful (#995, @chancez)
* Make auth verdicts visible in CLI (#1099, @meyskens)
* SCTP support (#977, @kaworu)
* UUID filter (#919, @kaworu)
* cmd/observe: improve help message for --first and --all options (#929, @rolinh)
* cmd/observe: improve help message for date formats of --since/--until (#956, @rolinh)
* cmd: Introduce `HUBBLE_COMPAT=legacy-json-output` (#865, @gandro)
* make: use `command -v` instead of `which` for better portability (#889, @rolinh)
* observe: traffic direction filter (#976, @kaworu)
* observe: warn on unknown field while JSON decoding (#962, @kaworu)

**Bugfixes:**
* Do not fail if unable to provide client certificate when requested (#996, @chancez)
* fix workload and identity filters (#1109, @kaworu)

**CI Changes:**
* .github: Add integration tests that installs cilium and queries it using Hubble CLI (#873, @chancez)
* .github: Configure renovate tag comment on GHA images (#1024, @chancez)
* Add golang to matchPackageNames for go deps groups (#1085, @chancez)
* Add unit tests for testing hubble args/flags handling (#874, @chancez)
* Configure Renovate (#1011, @renovate[bot])
* Fix Renovate datasources (#1029, @chancez)
* Remove dependabot configuration in favor of renovate (#1055, @chancez)
* Renovate: Ignore pflag (#1041, @chancez)
* Run renovate on Friday (#1056, @chancez)
* ci: enable new linters (#1064, @rolinh)
* ci: run codeql job on Ubuntu 22.04 and only on supported release branch (#968, @rolinh)
* dependabot: increase pull request limit and interval (#847, @kaworu)
* github: Enable dependabot for stable branch (#849, @gandro)
* make: add renovate anchor to the release target golang image (#1070, @kaworu)
* update Go to v1.20.2, golangci-lint to v1.52.2 (#967, @rolinh)

**Misc Changes:**
* .github: Replace deprecated command with environment file (#1049, @jongwooo)
* Add example links to renovate configuration in RELEASE.md (#1030, @chancez)
* all: bump Go to v1.20.3 (#981, @dependabot[bot])
* all: bump Go to v1.20.4 (#1008, @dependabot[bot])
* build: Bump golang image to alpine v3.18 (#1097, @kaworu)
* CHANGELOG: add links for each released version (#848, @kaworu)
* chore(deps): update actions/checkout action to v3.5.3 (master) (#1083, @renovate[bot])
* chore(deps): update actions/setup-go action to v4.0.1 (master) (#1034, @renovate[bot])
* chore(deps): update all github action dependencies (master) (minor) (#1079, @renovate[bot])
* chore(deps): update all github action dependencies (master) (minor) (#1092, @renovate[bot])
* chore(deps): update docker.io/library/alpine docker tag to v3.18.2 (master) (#1089, @renovate[bot])
* chore(deps): update docker.io/library/golang:1.20.5-alpine3.17 docker digest to eeac93e (master) (#1087, @renovate[bot])
* chore(deps): update docker/build-push-action action to v4.1.1 (master) (#1090, @renovate[bot])
* chore(deps): update docker/setup-buildx-action action to v2.8.0 (master) (#1105, @renovate[bot])
* chore(deps): update docker/setup-buildx-action action to v2.9.0 (master) (#1117, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.20.1 (master) (#1102, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.20.2 (master) (#1110, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.20.3 (master) (#1114, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.3.5 (master) (#1050, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.3.6 (master) (#1057, @renovate[bot])
* chore(deps): update golang to v1.20.5 (master) (patch) (#1066, @renovate[bot])
* chore(deps): update golangci/golangci-lint docker tag to v1.53.1 (master) (#1061, @renovate[bot])
* chore(deps): update golangci/golangci-lint docker tag to v1.53.2 (master) (#1077, @renovate[bot])
* chore(deps): update golangci/golangci-lint docker tag to v1.53.3 (master) (#1091, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v3.5.0 (master) (#1062, @renovate[bot])
* chore(deps): update helm/kind-action action to v1.6.0 (master) (#1037, @renovate[bot])
* chore(deps): update helm/kind-action action to v1.7.0 (master) (#1040, @renovate[bot])
* chore(deps): update library/golang docker tag to v1.20.5 (master) (#1071, @renovate[bot])
* chore(deps): update library/golang:1.20.5-alpine3.17 docker digest to eeac93e (master) (#1088, @renovate[bot])
* chore(deps): update skx/github-action-publish-binaries digest to 44887b2 (master) (#1022, @renovate[bot])
* ci: bump actions/checkout from 3.3.0 to 3.4.0 (#945, @dependabot[bot])
* ci: bump actions/checkout from 3.4.0 to 3.5.0 (#960, @dependabot[bot])
* ci: bump actions/checkout from 3.5.0 to 3.5.1 (#988, @dependabot[bot])
* ci: bump actions/checkout from 3.5.1 to 3.5.2 (#990, @dependabot[bot])
* ci: bump actions/download-artifact from 3.0.1 to 3.0.2 (#842, @dependabot[bot])
* ci: bump actions/setup-go from 3.5.0 to 4.0.0 (#943, @dependabot[bot])
* ci: bump docker/build-push-action from 3.2.0 to 3.3.0 (#852, @dependabot[bot])
* ci: bump docker/build-push-action from 3.3.0 to 4.0.0 (#884, @dependabot[bot])
* ci: bump docker/setup-buildx-action from 2.0.0 to 2.2.1 (#844, @dependabot[bot])
* ci: bump docker/setup-buildx-action from 2.2.1 to 2.3.0 (#882, @dependabot[bot])
* ci: bump docker/setup-buildx-action from 2.3.0 to 2.4.0 (#885, @dependabot[bot])
* ci: bump docker/setup-buildx-action from 2.4.0 to 2.4.1 (#891, @dependabot[bot])
* ci: bump docker/setup-buildx-action from 2.4.1 to 2.5.0 (#927, @dependabot[bot])
* ci: bump github/codeql-action from 2.1.37 to 2.1.38 (#850, @dependabot[bot])
* ci: bump github/codeql-action from 2.1.38 to 2.1.39 (#857, @dependabot[bot])
* ci: bump github/codeql-action from 2.1.39 to 2.2.0 (#877, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.0 to 2.2.1 (#881, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.1 to 2.2.2 (#890, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.11 to 2.2.12 (#991, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.12 to 2.3.0 (#997, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.2 to 2.2.3 (#896, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.3 to 2.2.4 (#898, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.4 to 2.2.5 (#915, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.5 to 2.2.6 (#926, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.6 to 2.2.7 (#944, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.7 to 2.2.8 (#954, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.8 to 2.2.9 (#963, @dependabot[bot])
* ci: bump github/codeql-action from 2.2.9 to 2.2.11 (#984, @dependabot[bot])
* ci: bump github/codeql-action from 2.3.0 to 2.3.2 (#1005, @dependabot[bot])
* ci: bump github/codeql-action from 2.3.2 to 2.3.3 (#1013, @dependabot[bot])
* ci: bump github/codeql-action from 2.3.3 to 2.3.4 (#1052, @dependabot[bot])
* ci: bump golangci/golangci-lint-action from 3.3.1 to 3.4.0 (#862, @dependabot[bot])
* dependabot: prefix backport PRs with target-branch version (#879, @kaworu)
* dockerfile: bump library/alpine from 3.17.1 to 3.17.2 (#902, @dependabot[bot])
* dockerfile: bump library/alpine from 3.17.2 to 3.17.3 (#971, @dependabot[bot])
* dockerfile: bump library/alpine from 3.17.3 to 3.18.0 (#1020, @dependabot[bot])
* dockerfile: bump library/alpine from `69665d0` to `ff6bdca` (#932, @dependabot[bot])
* dockerfile: bump library/golang from 1.19.5-alpine3.17 to 1.20.0-alpine3.17 (#888, @dependabot[bot])
* dockerfile: bump library/golang from 1.20.0-alpine3.17 to 1.20.1-alpine3.17 (#904, @dependabot[bot])
* dockerfile: bump library/golang from 1.20.1-alpine3.17 to 1.20.2-alpine3.17 (#921, @dependabot[bot])
* dockerfile: bump library/golang from `1db1276` to `576da1a` (#966, @dependabot[bot])
* dockerfile: bump library/golang from `1e29171` to `0d145ec` (#903, @dependabot[bot])
* dockerfile: bump library/golang from `48f336e` to `87d0a33` (#917, @dependabot[bot])
* dockerfile: bump library/golang from `4e6bc0e` to `1db1276` (#931, @dependabot[bot])
* dockerfile: bump library/golang from `576da1a` to `96a0a98` (#972, @dependabot[bot])
* dockerfile: bump library/golang from `96a0a98` to `87734b7` (#973, @dependabot[bot])
* fix(deps): pin dependencies (master) (#1032, @renovate[bot])
* fix(deps): pin dependencies (master) (#1045, @renovate[bot])
* fix(deps): update all go dependencies master (master) (minor) (#1093, @renovate[bot])
* fix(deps): update module github.com/sirupsen/logrus to v1.9.2 (master) (#1038, @renovate[bot])
* fix(deps): update module github.com/sirupsen/logrus to v1.9.3 (master) (#1078, @renovate[bot])
* fix(deps): update module github.com/spf13/cast to v1.5.1 (master) (#1033, @renovate[bot])
* fix(deps): update module github.com/stretchr/testify to v1.8.4 (master) (#1058, @renovate[bot])
* fix(deps): update module golang.org/x/sys to v0.10.0 (master) (#1111, @renovate[bot])
* fix(deps): update module google.golang.org/grpc to v1.56.1 (master) (#1103, @renovate[bot])
* fix(deps): update module google.golang.org/grpc to v1.56.2 (master) (#1115, @renovate[bot])
* fix(deps): update module google.golang.org/protobuf to v1.31.0 (master) (#1106, @renovate[bot])
* fix(deps): update module gopkg.in/yaml.v2 to v3 (master) (#1025, @renovate[bot])
* Newest stable release is Hubble v0.11.3 (#940, @gandro)
* Prepare for v0.12 development (#846, @gandro)
* README: Fix broken links (#909, @netoax)
* README: v0.11.1 is the latest release (#868, @gandro)
* Revert "Add hubble observe flows" (#869, @chancez)
* treewide: Bump newest release to v0.11.2 (#908, @gandro)
* Update doc and stable.txt for v0.11.5 release (#1018, @kaworu)
* Update doc and stable.txt for v0.11.6 release (#1075, @kaworu)
* Update release instructions (#994, @glibsm)
* Update things after 0.11.4 release (#1001, @glibsm)
* vendor: bump github.com/fatih/color from 1.13.0 to 1.14.0 (#860, @dependabot[bot])
* vendor: bump github.com/fatih/color from 1.14.0 to 1.14.1 (#863, @dependabot[bot])
* vendor: bump github.com/fatih/color from 1.14.1 to 1.15.0 (#928, @dependabot[bot])
* vendor: bump github.com/sirupsen/logrus from 1.9.0 to 1.9.1 (#1039, @dependabot[bot])
* vendor: bump github.com/spf13/cobra from 1.6.1 to 1.7.0 (#978, @dependabot[bot])
* vendor: bump github.com/spf13/viper from 1.14.0 to 1.15.0 (#859, @dependabot[bot])
* vendor: bump github.com/stretchr/testify from 1.8.1 to 1.8.2 (#914, @dependabot[bot])
* vendor: bump github.com/stretchr/testify from 1.8.2 to 1.8.3 (#1047, @dependabot[bot])
* vendor: bump github/cilium to v1.13.0-rc5 (#871, @rolinh)
* vendor: bump golang.org/x/net from 0.5.0 to 0.7.0 (#910, @dependabot[bot])
* vendor: bump golang.org/x/sys from 0.4.0 to 0.5.0 (#892, @dependabot[bot])
* vendor: bump golang.org/x/sys from 0.5.0 to 0.6.0 (#920, @dependabot[bot])
* vendor: bump golang.org/x/sys from 0.6.0 to 0.7.0 (#979, @dependabot[bot])
* vendor: bump golang.org/x/sys from 0.7.0 to 0.8.0 (#1014, @dependabot[bot])
* vendor: bump google.golang.org/grpc from 1.52.0 to 1.52.1 (#870, @dependabot[bot])
* vendor: bump google.golang.org/grpc from 1.52.1 to 1.52.3 (#876, @dependabot[bot])
* vendor: bump google.golang.org/grpc from 1.52.3 to 1.53.0 (#895, @dependabot[bot])
* vendor: bump google.golang.org/grpc from 1.53.0 to 1.54.0 (#953, @dependabot[bot])
* vendor: bump google.golang.org/grpc from 1.54.0 to 1.55.0 (#1015, @dependabot[bot])
* vendor: bump google.golang.org/protobuf from 1.28.1 to 1.29.0 (#924, @dependabot[bot])
* vendor: bump google.golang.org/protobuf from 1.29.0 to 1.29.1 (#939, @dependabot[bot])
* vendor: bump google.golang.org/protobuf from 1.29.1 to 1.30.0 (#949, @dependabot[bot])

## [v0.11.6] - 2023-06-07
[v0.11.6]: https://github.com/cilium/hubble/compare/v0.11.5...v0.11.6

**CI Changes:**
* [v0.11] .github: Configure renovate tag comment on GHA images (#1028, @chancez)
* [v0.11] Fix Renovate datasources (#1031, @chancez)
* [v0.11]: renovate go gha (#1069, @kaworu)
* make: add renovate anchor to the release target golang image (#1072, @kaworu)

**Misc Changes:**
* [v0.11] ci: Bump github/codeql-action from 2.3.3 to 2.3.4 (#1053, @dependabot[bot])
* chore(deps): update actions/setup-go action to v4.0.1 (v0.11) (#1035, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.3.5 (v0.11) (#1051, @renovate[bot])
* chore(deps): update github/codeql-action action to v2.3.6 (v0.11) (#1060, @renovate[bot])
* chore(deps): update golang docker tag to v1.19.10 (v0.11) (#1073, @renovate[bot])
* chore(deps): update golang stable to v1.19.10 (v0.11) (patch) (#1067, @renovate[bot])
* chore(deps): update golangci/golangci-lint-action action to v3.5.0 (v0.11) (#1063, @renovate[bot])
* chore(deps): update skx/github-action-publish-binaries digest to 44887b2 (v0.11) (#1026, @renovate[bot])

## [v0.11.5] - 2023-05-05
[v0.11.5]: https://github.com/cilium/hubble/compare/v0.11.4...v0.11.5

**Misc Changes:**
* [v0.11] ci: Bump github/codeql-action from 2.3.0 to 2.3.2 (#1006, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.3.2 to 2.3.3 (#1016, @dependabot[bot])
* v0.11/all: bump Go to v1.19.9 (#1007, @dependabot[bot])

## [v0.11.4] - 2023-04-24
[v0.11.4]: https://github.com/cilium/hubble/compare/v0.11.3...v0.11.4

**Misc Changes:**
* [v0.11] ci: Bump actions/checkout from 3.3.0 to 3.4.0 (#947, @dependabot[bot])
* [v0.11] ci: Bump actions/checkout from 3.4.0 to 3.5.0 (#961, @dependabot[bot])
* [v0.11] ci: Bump actions/checkout from 3.5.0 to 3.5.1 (#989, @dependabot[bot])
* [v0.11] ci: Bump actions/checkout from 3.5.1 to 3.5.2 (#992, @dependabot[bot])
* [v0.11] ci: Bump actions/setup-go from 3.5.0 to 4.0.0 (#946, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.11 to 2.2.12 (#993, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.12 to 2.3.0 (#998, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.6 to 2.2.7 (#948, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.7 to 2.2.8 (#955, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.8 to 2.2.9 (#964, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.9 to 2.2.11 (#985, @dependabot[bot])
* [v0.11] dockerfile: Bump library/alpine from 3.17.2 to 3.17.3 (#969, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from 1.19.7-alpine3.17 to 1.19.8-alpine3.17 (#980, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from `30630b1` to `31f980a` (#970, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from `31f980a` to `04065e6` (#974, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from `8b660f4` to `30630b1` (#965, @dependabot[bot])
* v0.11/vendor: bump cilium to v1.13.1 (#975, @rolinh)
* v0.11: bump Cilium to v1.13.2, update deps to their latest patch release (#999, @rolinh)

## [v0.11.3] - 2023-03-15
[v0.11.3]: https://github.com/cilium/hubble/compare/v0.11.2...v0.11.3

**Misc Changes:**
* [v0.11] ci: Bump docker/setup-buildx-action from 2.4.1 to 2.5.0 (#935, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.4 to 2.2.5 (#916, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.5 to 2.2.6 (#933, @dependabot[bot])
* [v0.11] dockerfile: Bump library/alpine from `69665d0` to `ff6bdca` (#934, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from 1.19.6-alpine3.17 to 1.19.7-alpine3.17 (#922, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from `31c62d9` to `62a2c84` (#918, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from `ee42797` to `8b660f4` (#936, @dependabot[bot])
* [v0.11] Update Golang to v1.19.7 (#930, @gandro)

## [v0.11.2] - 2023-02-15
[v0.11.2]: https://github.com/cilium/hubble/compare/v0.11.1...v0.11.2

**Misc Changes:**
* [v0.11] ci: Bump docker/setup-buildx-action from 2.2.1 to 2.3.0 (#883, @dependabot[bot])
* [v0.11] ci: Bump docker/setup-buildx-action from 2.3.0 to 2.4.0 (#886, @dependabot[bot])
* [v0.11] ci: Bump docker/setup-buildx-action from 2.4.0 to 2.4.1 (#893, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.0 to 2.2.1 (#880, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.1 to 2.2.2 (#894, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.2 to 2.2.3 (#897, @dependabot[bot])
* [v0.11] ci: Bump github/codeql-action from 2.2.3 to 2.2.4 (#899, @dependabot[bot])
* [v0.11] dockerfile: Bump library/alpine from 3.17.1 to 3.17.2 (#901, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from 1.19.5-alpine3.17 to 1.19.6-alpine3.17 (#905, @dependabot[bot])
* [v0.11] dockerfile: Bump library/golang from `2381c1e` to `a00a03c` (#900, @dependabot[bot])
* ci: bump github/codeql-action from 2.1.39 to 2.2.0 (#878, @dependabot[bot])
* Update Go to v1.19.6 (#906, @gandro)
* v0.11: vendor: bump github/cilium to v1.13.0-rc5 (#872, @rolinh)

**Other Changes:**
* [v0.11] ci: Bump docker/build-push-action from 3.3.0 to 4.0.0 (#887, @dependabot[bot])


## [v0.11.1] - 2023-01-24
[v0.11.1]: https://github.com/cilium/hubble/compare/v0.11.0...v0.11.1

**Minor Changes:**
* [v0.11] cmd: Introduce `HUBBLE_COMPAT=legacy-json-output` (#866, @gandro)

**Misc Changes:**
* ci: bump actions/download-artifact from 3.0.1 to 3.0.2 (#855, @dependabot[bot])
* ci: bump docker/build-push-action from 3.2.0 to 3.3.0 (#854, @dependabot[bot])
* ci: bump docker/setup-buildx-action from 2.0.0 to 2.2.1 (#856, @dependabot[bot])
* ci: bump github/codeql-action from 2.1.37 to 2.1.38 (#853, @dependabot[bot])
* ci: bump github/codeql-action from 2.1.38 to 2.1.39 (#858, @dependabot[bot])
* ci: bump golangci/golangci-lint-action from 3.3.1 to 3.4.0 (#864, @dependabot[bot])

## [v0.11.0] - 2023-01-11
[v0.11.0]: https://github.com/cilium/hubble/compare/v0.10.0...v0.11.0

This v0.11.0 release of the Hubble CLI adds support for features added in
Cilium v1.13: Hubble now has visibility into Cilium's SockLB,
meaning it is possible to observe service address translations performed
by Cilium on the socket level (#816). Hubble CLI v0.11 also supports the newly
introduced Cilium v1.13 flow filters for workload and trace ID (#794, #795).
Another noteworthy change is the newly displayed traffic direction for
policy verdict events in the `-o compact` output (#759).

*Breaking Changes*

In accordance with semver 0.x releases, this release contains a breaking change
to the Hubble command-line output:

 - This release also removes the old and deprecated JSON formatter and now always
   uses to the more flexible proto3-based `jsonpb` output when JSON is selected
   as the output format. This is a potentially breaking change and requires that
   e.g. `jq` queries of the form `hubble observe -o json | jq .source` are
   rewritten as `hubble observe -o json | jq .flow.source` (#826).

**Major Changes:**
* Add support for SockLB events (#816, @gandro)
* cmd: Make `-o json` an alias for `-o jsobpb` (#826, @gandro)

**Minor Changes:**
* Add endpoint workload filters (#794, @chancez)
* Add traceID filter (#795, @chancez)
* compact: Add traffic direction to policy verdict events (#759, @michi-covalent)

**Bugfixes:**
* cmd/observe: fix stdin reading from file redirection (#815, @rolinh)

**CI Changes:**
* ci: update golangci-lint config, add new linters (#814, @rolinh)
* dependabot config improvements (#836, @kaworu)
* Makefile: Fix potential uid/gid collision by using setpriv (#821, @gandro)

**Misc Changes:**
* Add Code of Conduct (#828, @xmulligan)
* build(deps): bump actions/checkout from 3.0.2 to 3.1.0 (#799, @dependabot[bot])
* build(deps): bump actions/download-artifact from 3.0.0 to 3.0.1 (#820, @dependabot[bot])
* build(deps): bump actions/setup-go from 3.2.0 to 3.2.1 (#765, @dependabot[bot])
* build(deps): bump actions/setup-go from 3.2.1 to 3.3.0 (#783, @dependabot[bot])
* build(deps): bump actions/setup-go from 3.3.0 to 3.3.1 (#813, @dependabot[bot])
* build(deps): bump actions/setup-go from 3.3.1 to 3.5.0 (#829, @dependabot[bot])
* build(deps): bump actions/upload-artifact from 3.1.0 to 3.1.2 (#834, @dependabot[bot])
* build(deps): bump docker/build-push-action from 3.0.0 to 3.1.0 (#770, @dependabot[bot])
* build(deps): bump docker/build-push-action from 3.1.0 to 3.1.1 (#775, @dependabot[bot])
* build(deps): bump docker/build-push-action from 3.1.1 to 3.2.0 (#801, @dependabot[bot])
* build(deps): bump docker/login-action from 2.0.0 to 2.1.0 (#805, @dependabot[bot])
* build(deps): bump github.com/google/go-cmp from 0.5.8 to 0.5.9 (#788, @dependabot[bot])
* build(deps): bump github.com/sirupsen/logrus from 1.8.1 to 1.9.0 (#766, @dependabot[bot])
* build(deps): bump github.com/spf13/cobra from 1.5.0 to 1.6.1 (#806, @dependabot[bot])
* build(deps): bump github.com/spf13/viper from 1.12.0 to 1.13.0 (#790, @dependabot[bot])
* build(deps): bump github.com/spf13/viper from 1.13.0 to 1.14.0 (#808, @dependabot[bot])
* build(deps): bump github.com/stretchr/testify from 1.7.3 to 1.7.5 (#756, @dependabot[bot])
* build(deps): bump github.com/stretchr/testify from 1.7.5 to 1.8.0 (#758, @dependabot[bot])
* build(deps): bump github.com/stretchr/testify from 1.8.0 to 1.8.1 (#802, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.12 to 2.1.14 (#755, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.14 to 2.1.15 (#757, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.15 to 2.1.16 (#762, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.16 to 2.1.18 (#777, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.18 to 2.1.19 (#778, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.19 to 2.1.22 (#785, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.22 to 2.1.24 (#789, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.24 to 2.1.25 (#791, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.25 to 2.1.26 (#793, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.26 to 2.1.27 (#796, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.27 to 2.1.35 (#822, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.35 to 2.1.36 (#824, @dependabot[bot])
* build(deps): bump github/codeql-action from 2.1.36 to 2.1.37 (#825, @dependabot[bot])
* build(deps): bump golang.org/x/sys from 0.2.0 to 0.3.0 (#823, @dependabot[bot])
* build(deps): bump golang.org/x/sys from 0.3.0 to 0.4.0 (#835, @dependabot[bot])
* build(deps): bump golangci/golangci-lint-action from 3.2.0 to 3.3.0 (#807, @dependabot[bot])
* build(deps): bump golangci/golangci-lint-action from 3.3.0 to 3.3.1 (#818, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.47.0 to 1.48.0 (#763, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.48.0 to 1.49.0 (#784, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.49.0 to 1.50.0 (#797, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.50.0 to 1.50.1 (#800, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.50.1 to 1.51.0 (#819, @dependabot[bot])
* build(deps): bump google.golang.org/protobuf from 1.28.0 to 1.28.1 (#771, @dependabot[bot])
* CHANGELOG.md: fix PR ref in the v0.10.0 release note (#753, @kaworu)
* ci: add new linters (#780, @tklauser)
* ci: bump actions/checkout from 3.1.0 to 3.3.0 (#838, @dependabot[bot])
* cmd/observe: stop sorting reserved identity names (#798, @kaworu)
* CODEOWNERS: update teams following removal of non-sig teams (#767, @tklauser)
* compact: Include DNS observation source (#803, @michi-covalent)
* Convert to SPDX license headers and remove copyright year (#812, @tklauser)
* dockerfile: bump library/alpine from 3.17.0 to 3.17.1 (#837, @dependabot[bot])
* dockerfile: bump library/golang from 1.19.4-alpine3.17 to 1.19.5-alpine3.17 (#840, @dependabot[bot])
* Fix observe command not supporting --until without --since (#792, @ChrsMark)
* Link to release v0.10.0 (#750, @gandro)
* Makefile: Run release build as regular user (#751, @gandro)
* Update Go to 1.18.4 (#760, @tklauser)
* Update Go to 1.18.5 (#772, @tklauser)
* Update Go to 1.19 (#779, @tklauser)
* Update Go to 1.19.3, golangci-lint to 1.50.1 (#810, @tklauser)
* Update Go to v1.19.4 and alpine to v3.17.0 (#832, @kaworu)
* update Go to v1.19.5 (#841, @rolinh)
* Use command path when registering flagsets (#769, @chancez)
* vendor: Bump Cilium to v1.13 branch (#843, @gandro)
* vendor: bump google.golang.org/grpc from 1.51.0 to 1.52.0 (#839, @dependabot[bot])

## [v0.10.0] - 2022-06-22
[v0.10.0]: https://github.com/cilium/hubble/compare/v0.9.0...v0.10.0

The v0.10.0 release of the Hubble CLI coincides with Cilium v1.12.
It adds a new `--first` option to query for earlier flows and events
(#719, requires Cilium v1.12 and newer), further improves the default `compact`
output by displaying security identities and refining policy verdict event output
(#717, #734, #745), and deprecates the `-o json` option in favor of `-o jsonpb`
(#738).

This release also contains many quality of life improvements, such as more
flexible time range filter parsing (#707), extended shell completion for
various filter flags (#727, #744), support for named identity filters (#732),
improvements to the command-line usage documentation (#718, #730, #731, #733),
and an updated version of the Hubble logo (#726).

**Major Changes:**
* cli: Deprecate `-o json`, recommend `-o jsonpb` instead (#738, @gandro)
* cmd/observe: Add `--first` to support querying for earlier flows and events (#719, @chancez)
* printer: Display security identity in compact output (#717, @gandro)

**Minor Changes:**
* Add support for less granular time formats (#707, @rolinh)
* cmd/observe: add flag completion for `--protocol` (#727, @rolinh)
* cmd/observe: document subtypes and add completion for subtypes (#744, @rolinh)
* cmd/observe: improve policy verdict output in compact mode (#745, @rolinh)

**Bugfixes:**
* cmd/config: ensure that the configuration directory exist (#684, @rolinh)
* cmd/observe: match only Hubble-specific part of error in Test_getFlowsRequestWithInvalidRawFilters (#655, @tklauser)

**CI Changes:**
* .github: let dependabot ignore Cilium dependency (#675, @tklauser)

**Misc Changes:**
* build(deps): bump actions/checkout from 2.4.0 to 3 (#693, @dependabot[bot])
* build(deps): bump actions/checkout from 3.0.0 to 3.0.1 (#705, @dependabot[bot])
* build(deps): bump actions/checkout from 3.0.1 to 3.0.2 (#709, @dependabot[bot])
* build(deps): bump actions/download-artifact from 2.0.10 to 2.1.0 (#668, @dependabot[bot])
* build(deps): bump actions/download-artifact from 2.1.0 to 3 (#688, @dependabot[bot])
* build(deps): bump actions/setup-go from 2.1.4 to 2.1.5 (#665, @dependabot[bot])
* build(deps): bump actions/setup-go from 2.1.5 to 2.2.0 (#680, @dependabot[bot])
* build(deps): bump actions/setup-go from 2.2.0 to 3 (#697, @dependabot[bot])
* build(deps): bump actions/setup-go from 3.1.0 to 3.2.0 (#746, @dependabot[bot])
* build(deps): bump actions/upload-artifact from 2.2.4 to 2.3.0 (#662, @dependabot[bot])
* build(deps): bump actions/upload-artifact from 2.3.0 to 2.3.1 (#663, @dependabot[bot])
* build(deps): bump actions/upload-artifact from 2.3.1 to 3 (#701, @dependabot[bot])
* build(deps): bump actions/upload-artifact from 3.0.0 to 3.1.0 (#724, @dependabot[bot])
* build(deps): bump docker/build-push-action from 2.10.0 to 3 (#728, @dependabot[bot])
* build(deps): bump docker/build-push-action from 2.7.0 to 2.8.0 (#673, @dependabot[bot])
* build(deps): bump docker/build-push-action from 2.8.0 to 2.9.0 (#679, @dependabot[bot])
* build(deps): bump docker/build-push-action from 2.9.0 to 2.10.0 (#699, @dependabot[bot])
* build(deps): bump docker/login-action from 1.10.0 to 1.12.0 (#669, @dependabot[bot])
* build(deps): bump docker/login-action from 1.12.0 to 1.13.0 (#683, @dependabot[bot])
* build(deps): bump docker/login-action from 1.13.0 to 1.14.1 (#704, @dependabot[bot])
* build(deps): bump docker/login-action from 1.14.1 to 2 (#742, @dependabot[bot])
* build(deps): bump docker/setup-buildx-action from 1.6.0 to 2 (#714, @dependabot[bot])
* build(deps): bump github.com/cilium/cilium from 1.11.0 to 1.11.1 (#674, @dependabot[bot])
* build(deps): bump github.com/google/go-cmp from 0.5.6 to 0.5.7 (#676, @dependabot[bot])
* build(deps): bump github.com/google/go-cmp from 0.5.7 to 0.5.8 (#712, @dependabot[bot])
* build(deps): bump github.com/spf13/cast from 1.4.1 to 1.5.0 (#725, @dependabot[bot])
* build(deps): bump github.com/spf13/cobra from 1.2.1 to 1.3.0 (#664, @dependabot[bot])
* build(deps): bump github.com/spf13/cobra from 1.3.0 to 1.4.0 (#694, @dependabot[bot])
* build(deps): bump github.com/spf13/viper from 1.10.0 to 1.10.1 (#667, @dependabot[bot])
* build(deps): bump github.com/spf13/viper from 1.10.1 to 1.11.0 (#706, @dependabot[bot])
* build(deps): bump github.com/spf13/viper from 1.11.0 to 1.12.0 (#729, @dependabot[bot])
* build(deps): bump github.com/stretchr/testify from 1.7.0 to 1.7.1 (#698, @dependabot[bot])
* build(deps): bump github.com/stretchr/testify from 1.7.1 to 1.7.2 (#743, @dependabot[bot])
* build(deps): bump github/codeql-action from 1 to 2 (#711, @dependabot[bot])
* build(deps): bump github/codeql-action from 96bc9c36c68e097cd033777efed25c248ffcf09a to 2.1.12 (#735, @dependabot[bot])
* build(deps): bump golangci/golangci-lint-action from 2 to 3.1.0 (#685, @dependabot[bot])
* build(deps): bump golangci/golangci-lint-action from 3.1.0 to 3.2.0 (#720, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.42.0 to 1.43.0 (#666, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.43.0 to 1.44.0 (#678, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.44.0 to 1.45.0 (#702, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.45.0 to 1.46.0 (#710, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.46.0 to 1.46.2 (#721, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.46.2 to 1.47.0 (#736, @dependabot[bot])
* build(deps): bump google.golang.org/protobuf from 1.27.1 to 1.28.0 (#700, @dependabot[bot])
* bump Go to v1.18.1, update golangci-lint to v1.45.2 (#708, @rolinh)
* ci: bump golangci-lint to v1.45.0 (#696, @rolinh)
* ci: use hashes for all GitHub Action modules (#722, @rolinh)
* compact: Use "ID" for security identity prefix (#734, @michi-covalent)
* Dockerfile: fix golang image name to v1.18.2 (#723, @kaworu)
* docs: Document pod/service filter prefix behavior (#733, @slayer321)
* docs: update logos and add dark logo (#726, @raphink)
* docs: update the cli doc with cidr range source/destination ip filter (#731, @slayer321)
* go.mod, vendor: update cilium to 1.11.0 (#658, @tklauser)
* improve cli help text for service filtering (#730, @ILLIDOM)
* named reserved identites support for `--{,from-,to-}identity` (#732, @kaworu)
* Prepare for v0.10 development cycle (#652, @gandro)
* Refactor usage template to determine --help flags using a registration pattern (#718, @chancez)
* release and changelog misc improvements (#659, @kaworu)
* Update Cobra to v1.5.0 (#747, @rolinh)
* Update Go to 1.17.4 and alpine to 3.15 (#653, @tklauser)
* Update Go to 1.17.5 (#660, @tklauser)
* Update Go to 1.17.6 (#670, @tklauser)
* Update Go to 1.17.7 (#681, @tklauser)
* Update Go to 1.17.8 (#689, @tklauser)
* Update Go to 1.18.2 (#715, @tklauser)
* Update Go to 1.18.3, alpine to 3.16, golangci-lint to 1.46.2 (#737, @tklauser)
* Update Go to v1.18 (#695, @rolinh)
* vendor: Bump Cilium to v1.12 branch (#748, @gandro)
* vendor: update yaml.v3 to v3.0.1 (#741, @kaworu)

## [v0.9.0] - 2021-11-30
[v0.9.0]: https://github.com/cilium/hubble/compare/v0.8.2...v0.9.0

Hubble v0.9.0 coincides with Cilium v1.11. It brings many improvements to the
CLI: Colored output (#551), improved readability and alternative output formats
in `hubble status` (#629, #614), and the ability to specify custom filters via
the newly introduced `--allowlist` and `--denylist` flags (#643). Other changes
include automatic stop conditions for `hubble record` (#607), omit displaying
old flows in follow mode by default (#573) and client binary support for
Windows ARM64 (#618).

**Minor Changes:**
* build release binaries for Windows ARM64 (#618, @rolinh)
* cmd/observe: add color support (#551, @rolinh)
* cmd/observe: do not set `--last` to 20 by default in follow mode (#573, @rolinh)
* cmd/record: Add stop condition flags (#607, @gandro)
* cmd/status: add support for multiple output formats (#614, @rolinh)
* observe: Add --allowlist / --denylist flags (#643, @michi-covalent)
* printer: group digits by 3 for flow counters and make uptime human-readable (#629, @rolinh)
* Update cobra to v1.2.1 and use built-in completion command (#582, @rolinh)

**Bugfixes:**
* printer: Add missing verdicts (#626, @pchaigno)
* printer: fix dict outout newline (#615, @rolinh)

**CI Changes:**
* .github: Cancel outdated PR and push workflows (#555, @pchaigno)
* Add CODEOWNERS (#576, @gandro)
* ci: bump golangci-lint to v1.42.0 (#611, @tklauser)
* CODEOWNERS: assign GH actions to github-sec team (#577, @tklauser)

**Misc Changes:**
* .github/workflows: move Go module vendoring check to build checks (#563, @tklauser)
* .github: Rename maintainer's little helper's config file (#569, @pchaigno)
* build(deps): bump actions/checkout from 2 to 2.3.5 (#640, @dependabot[bot])
* build(deps): bump actions/checkout from 2.3.5 to 2.4.0 (#648, @dependabot[bot])
* build(deps): bump actions/setup-go from 2.1.3 to 2.1.4 (#616, @dependabot[bot])
* build(deps): bump github.com/fatih/color from 1.10.0 to 1.12.0 (#558, @dependabot[bot])
* build(deps): bump github.com/fatih/color from 1.12.0 to 1.13.0 (#633, @dependabot[bot])
* build(deps): bump github.com/google/go-cmp from 0.5.5 to 0.5.6 (#561, @dependabot[bot])
* build(deps): bump github.com/spf13/cast from 1.3.1 to 1.4.0 (#600, @dependabot[bot])
* build(deps): bump github.com/spf13/cast from 1.4.0 to 1.4.1 (#613, @dependabot[bot])
* build(deps): bump github.com/spf13/viper from 1.8.0 to 1.8.1 (#579, @dependabot[bot])
* build(deps): bump github.com/spf13/viper from 1.8.1 to 1.9.0 (#628, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.38.0 to 1.39.0 (#584, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.39.0 to 1.39.1 (#608, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.39.1 to 1.40.0 (#610, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.40.0 to 1.41.0 (#634, @dependabot[bot])
* build(deps): bump google.golang.org/grpc from 1.41.0 to 1.42.0 (#649, @dependabot[bot])
* build(deps): bump google.golang.org/protobuf from 1.26.0 to 1.27.1 (#583, @dependabot[bot])
* build(deps): bump skx/github-action-publish-binaries from c881a3f8ffb80b684f367660178d38ceabc065c2 to 2.0 (#632, @dependabot[bot])
* ci: bump Go to 1.17 for golangci-lint (#625, @kaworu)
* ci: enable checks for missing Go documentation (#581, @rolinh)
* ci: fix the go vendoring check (#575, @kaworu)
* cmd/observe: move flows specific code to flows*.go files (#578, @tklauser)
* cmd/observe: remove depreacted formatting flags (json, compact, dict) (#603, @rolinh)
* cmd/observe: rename flow-related functions, types and import aliases (#574, @tklauser)
* CODEOWNERS: assign Go module vendoring to @cilium/vendor (#580, @tklauser)
* docker: add note about bogus busybox's nslookup implementation (#587, @rolinh)
* Fix broken link (#593, @sharjeelaziz)
* git commands in the Makefile return the empty string if they fail. (#589, @zhiyanfoo)
* github: Add "Image Release Build" workflow (#627, @gandro)
* Log a debug message when reading from stdin (#598, @michi-covalent)
* Makefile: Introduce GO_BUILD variable (#560, @gandro)
* pkg/printer: disable color output in tests (#562, @tklauser)
* Prepare for 0.9 development cycle (#545, @gandro)
* readme: clarify that only the latest version is maintained/supported (#568, @rolinh)
* RELEASE.md: document Homebrew formular update as optional step (#624, @tklauser)
* RELEASE.md: fix brew command formatting (#630, @tklauser)
* release: pin skx/github-action-publish-binaries to a specific sha (#546, @rolinh)
* Small test cleanups (#571, @tklauser)
* Update doc and stable.txt for v0.8.2 release (#623, @rolinh)
* Update Go to 1.16.4 (#548, @tklauser)
* Update Go to 1.16.5 (#564, @tklauser)
* Update Go to 1.16.7 (#604, @tklauser)
* Update Go to 1.17 (#612, @tklauser)
* Update Go to 1.17.2 (#635, @tklauser)
* update Go to v1.16.6, alpine to 3.14 (#585, @rolinh)
* update Go to v1.17.1 (#620, @rolinh)
* Update Go to v1.17.3, golangci-lint to v1.43.0 (#646, @rolinh)
* Update readme, changelog and stable.txt for v0.8.1 release (#595, @rolinh)
* Use golangci-lint for static checks (#559, @rolinh)
* vendor: bump github.com/cilium/cilium to latest master (#556, @tklauser)
* vendor: Bump github.com/cilium/cilium to v1.11.0-rc3 (#650, @gandro)
* vendor: bump grpc to v1.37.1; honnef tools to v0.1.4 (#552, @rolinh)
* vendor: bump viper to v1.8.0 and grpc to v1.38.0 (#572, @rolinh)
* version: Drop the "v" prefix (#638, @michi-covalent)

## [v0.8.2] - 2021-09-10
[v0.8.2]: https://github.com/cilium/hubble/compare/v0.8.1...v0.8.2

This patch release fixes a bug in the dict output where a newline was missing.
It also removes long-deprecated `--json`, `--compact` and `--dict` flags (use
the `--output` flag instead) that actually turned out to be broken at this
point. A very visible addition is color support, a change that was backported
from the `master` branch upon popular request.
In addition, the Go version, which is used to create release binaries, is
updated to the latest v1.16.8 and the Cilium dependency is updated to v1.10.4.

**Minor Changes:**
* Backport color output to v0.8 branch (#609, @michi-covalent)

**Bugfixes:**
* v0.8: printer: fix dict outout newline (#617, @rolinh)

**Misc Changes:**
* [v0.8] go.mod, vendor: bump cilium to v1.10.4 (#619, @tklauser)
* v0.8: cmd/observe: remove depreacted formatting flags (json, compact, dict) (#606, @rolinh)
* v0.8: Update Go to 1.16.7 (#605, @tklauser)
* v0.8: update Go to v1.16.8, Alpine base image to 3.14.2 (#621, @rolinh)

## [v0.8.1] - 2021-07-19
[v0.8.1]: https://github.com/cilium/hubble/compare/v0.8.0...v0.8.1

This patch release updates gRPC and Cilium dependencies to v1.37.1 and v1.10.3
respectively. The Go version, which is used to create release binaries, is also
updated to the latest v1.16.6. A minor, mostly cosmetic, bug is also fixed
which allows building Hubble without any warning being displayed when the `.git`
directory is not present.

**Misc Changes:**

* v0.8: bump cilium to v1.10.2, Go to v1.16.6 (#586, @rolinh)
* v0.8: release: pin skx/github-action-publish-binaries to a specific sha (#547, @rolinh)
* v0.8: update cilium to v1.10.3, backport git version fix (#591, @rolinh)
* v0.8: Update Go to 1.16.4 (#549, @tklauser)
* v0.8: Update Go to 1.16.5 (#565, @tklauser)
* v0.8: vendor: bump github.com/cilium/cilium to v1.10.0 (#557, @tklauser)
* v0.8: vendor: bump grpc to v1.37.1; honnef tools to v0.1.4 (#553, @rolinh)

## [v0.8.0] - 2021-05-03
[v0.8.0]: https://github.com/cilium/hubble/compare/v0.7.1...v0.8.0

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

## [v0.7.1] - 2020-10-22
[v0.7.1]: https://github.com/cilium/hubble/compare/v0.7.0...v0.7.1

**Bugfixes:**
* cmd: fix environment variable names for options with dashes (#408, @Rolinh)

**Misc Changes:**
* build: ensure that binaries are always statically built (#402, @Rolinh)

## [v0.7.0] - 2020-10-19
[v0.7.0]: https://github.com/cilium/hubble/compare/v0.6.1...v0.7.0

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

## [v0.6.1] - 2020-06-12
[v0.6.1]: https://github.com/cilium/hubble/compare/v0.6.0...v0.6.1

**Bugfixes:**
* make: fix git hash variable assignments for old make versions (#291, @michi-covalent)

**Misc Changes:**
* update Go version to v1.14.4 and alpine base image to v3.12 (#280, @Rolinh)

**Other Changes:**
* Backport https://github.com/cilium/hubble/pull/285 (#286, @michi-covalent)
* Prepare v0.6.1 release (#289, @michi-covalent)
* Require Cilium 1.7.x (#287, @michi-covalent)

## [v0.6.0] - 2020-05-29
[v0.6.0]: https://github.com/cilium/hubble/compare/v0.5.0...v0.6.0

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

## v0.5.0 - 2020-03-23
