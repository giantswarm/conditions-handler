# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

## [0.3.0] - 2022-03-31

### Changed

- Update github.com/giantswarm/conditions to v0.5.0
- Update github.com/giantswarm/microerror to v0.4.0
- Update github.com/giantswarm/micrologger to v0.6.0
- Update k8s.io/api to v0.22.2
- Update k8s.io/apimachinery to v0.22.2
- Update sigs.k8s.io/cluster-api to v1.0.5
- Update sigs.k8s.io/controller-runtime to v0.10.3
- Update sigs.k8s.io/yaml to v1.3.0
- Update architect to v4.15.0

## [0.2.1] - 2021-01-27

### Fixed

- `MachinePool` `ReplicasReady` is not always true in clusters with cluster-autoscaler enabled.

## [0.2.0] - 2021-01-11

### Added

- New handler that is setting `MachinePool` `ReplicasReady` condition.

### Changed

- `MachinePool` `Ready` condition is now summarizing `ReplicasReady` and `InfrastructureReady`, so both Kubernetes nodes and Azure infrastructure are taken into account.
- Added new `ReplicasReady` condition handler to the default `MachinePool` composite handler that is used in `azure-operator`.

## [0.1.2] - 2020-12-22

### Changed

- Simplified `InfrastructureReady` condition update, not checking object age and
  using only severity `Warning` when there is an issue.
- Simplified `ControlPlaneReady` condition update, not checking object age and
  using only severity `Warning` when there is an issue.

## [0.1.1] - 2020-12-11

### Changed

- Bump `github.com/giantswarm/micrologger` to `v0.4.0` 
- Use new logger functions `Debugf` and `Errorf`.
- Handle API conflict errors.

## [0.1.0] - 2020-12-08

- Generic condition handler for summarizing conditions into a single summary condition.
- Generic composite condition handlers for combining multiple handlers into one.
- Condition handlers for generic `Creating`, `Upgrading`, `Ready` and `InfrastructureReady` conditions that can be used in `Cluster` and `MachinePool` controllers.
- Condition handlers for `Cluster` `ControlPlaneReady` and `NodePoolsReady` conditions.
- Factory functions for creating `Cluster` and `MachinePool` conditions handlers that can be then used in an operator out of the box.

[Unreleased]: https://github.com/giantswarm/conditions-handler/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/giantswarm/conditions-handler/compare/v0.2.1...v0.3.0
[0.2.1]: https://github.com/giantswarm/conditions-handler/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/giantswarm/conditions-handler/compare/v0.1.2...v0.2.0
[0.1.2]: https://github.com/giantswarm/conditions-handler/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/giantswarm/conditions-handler/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/giantswarm/conditions-handler/releases/tag/v0.1.0
