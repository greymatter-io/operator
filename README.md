# Grey Matter Operator

Grey Matter Operator is a Kubernetes operator that enables using `greymatter.io/v1alpha1.Mesh` custom resource objects to manage Grey Matter Mesh deployments in a Kubernetes cluster.

This project is currently in an unstable alpha stage. This README will be updated constantly throughout the initial development process. For now, most things documented here will be in short-form.

## Development

### Requirements

This project is built using the [Operator SDK](https://sdk.operatorframework.io) which relies on [Kubebuilder](https://kubebuilder.io) for its CLI and the [controller-runtime project](https://github.com/kubernetes-sigs/controller-runtime). As such, this project adheres to conventions and best practices recommended by those projects, and uses the [Operator SDK CLI v1.12.0](https://sdk.operatorframework.io/docs/installation/) for maintaining its API, manifests, and general project structure.

