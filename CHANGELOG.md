# Changelog

## vNEXT

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
