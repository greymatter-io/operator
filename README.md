# Grey Matter Operator

Grey Matter Operator is a Kubernetes operator that enables using `greymatter.io/v1alpha1.Mesh` custom resource objects to manage Grey Matter Mesh deployments in a Kubernetes cluster.

This project is currently in an unstable alpha stage. This README will be updated constantly throughout the initial development process. For now, most things documented here will be in short-form.

## Development

### Requirements

This project is built on the [Operator SDK](https://sdk.operatorframework.io) which relies on [Kubebuilder](https://kubebuilder.io) for its CLI and the [controller-runtime project](https://github.com/kubernetes-sigs/controller-runtime). As such, this project adheres to conventions and best practices recommended by those projects.

Download the [Operator SDK CLI v1.12.0](https://sdk.operatorframework.io/docs/installation/) for maintaining the project API, manifests, and general project structure.

All other dependencies for this project can be added with `go mod vendor`, plus additional `make` targets which will download binaries to the `bin` directory.

### Local Quickstart

This section outlines how to set up a local development environment in [K3d](https://k3d.io) (not a requirement for this project, but an alternative to deploying to an online Kubernetes cluster).

The following commands set up a local cluster, install necessary resources in it, builds a binary, and runs it from outside of the cluster on your host.

```
k3d cluster create gm-operator -a 1 -p 30000:10808@loadbalancer
export KUBECONFIG=$(k3d kubeconfig write gm-operator)
make install run
```

The `make run` target runs the built binary.

To uninstall and tear down the local cluster, exit the terminal process and run:

```
make uninstall
k3d cluster delete gm-operator
```

### Deployed Quickstart

Warning: This setup is only for development! Do not use in production.

`make dev-run` - Builds an image, pushes it to `docker.greymatter.io/internal`, and installs to an OpenShift cluster to be managed by the Operator Lifecycle Manager.
`make dev-cleanup` - Remove from the OpenShift cluster.
