# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).



## [Unreleased]

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

[Unreleased]: https://github.com/giantswarm/conditions-handler/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/giantswarm/conditions-handler/releases/tag/v0.1.0
