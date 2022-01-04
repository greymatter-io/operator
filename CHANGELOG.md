# Changelog

## 0.3.0 (January 4, 2022)

### Added

- Add SPIRE as SPIFFE implementation for mutual TLS between workloads

### Fixed

- Prevent `greymatter apply` commands to Control API from running until it has connected to Redis.

## 0.2.0 (December 10, 2021)

### Changed

- Upgrade Grey Matter CLI binary dependency from 3.0.0 to 4.0.1

## 0.1.2 (December 2, 2021)

### Changed

- Change `GREYMATTER_DOCKER_*` config env vars to `GREYMATTER_REGISTRY_*`
- Change k8s-operator `--username` and `--password` flags to `--registry-username`
  and `--registry-password`, respectively. And remove -u and -p aliases to
  reserve these for future use.

## 0.1.0 (December 1, 2021)

This is a pre-release with basic support for installing Grey Matter core 
components and dependencies and bootstrapping Grey Matter mesh configurations.

### Added

- Support for general Kubernetes distributions
- Support for OpenShift, packaged for compatibility with the Operator Lifecycle Manager
