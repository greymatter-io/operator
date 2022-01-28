# Changelog

## NEXT

### Added

- [greymatter-cue](https://github.com/greymatter-io/greymatter-cue) Git submodule for importing
  Grey Matter CUE schemas.
- Generated CUE schema for the Mesh custom resource definition, to be parsed within CUE when
  generating installation manifests and mesh configurations for a configured Mesh.

### Changed

- Moved all CUE packages into a single CUE module that lives in `pkg/cuemodule`. Each CUE package
  may be evaluated using the CUE CLI as well as loaded via exported loader functions defined in the
  `cuemodule` Go package.

### Fixed

- Addressed a bug that kept existing Mesh custom resources from being fully registered with a new
  instance of the operator on startup. The bug prevented new mesh configurations from being
  generated for workloads within namespaces of existing meshes.

## 0.3.0 (January 10, 2022)

### Added

- SPIRE as SPIFFE implementation for mutual TLS between workloads

### Fixed

- Prevent `greymatter apply` commands to Control API from running until it has connected to Redis.

### Removed

- Support for an external Redis cache for mesh configurations, opting for an internally-managed one
  secured by mutual TLS (via SPIRE).
- Catalog entry for Redis
- JWT Security service's Redis dependency
- Listeners using port 10707 for initial boootstrapping of mesh configuration

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
