# Grey Matter Operator

Grey Matter Operator is a Kubernetes operator that enables using `greymatter.io/v1alpha1.Mesh` custom resource objects to manage Grey Matter Mesh deployments in a Kubernetes cluster.

This project is currently in an unstable alpha stage. This README will be updated constantly throughout the initial development process. For now, most things documented here will be in short-form.

## Development

### Dependencies

Grey Matter Operator is built with [Go 1.7](https://golang.org/dl/) using the [Operator SDK](https://sdk.operatorframework.io). As such, this project adheres to its conventions and recommended best practices.

Download the [Operator SDK CLI v1.12.0](https://sdk.operatorframework.io/docs/installation/) for maintaining the project API, manifests, and general project structure.

It also uses [Cue](https://cuelang.org/docs/install/) for maintaining much of its internal API.

If building for [OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift/container-platform), you'll also need the [OpenShift CLI](https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/).

Lastly, while not absolutely necessary, you may need [cfssl](https://github.com/cloudflare/cfssl) for generating certs used by the operator's webhook server when developing against non-Openshift environments. Run `go get github.com/cloudflare/cfssl/cmd/...` to download all binaries to the `bin` directory of your `$GOPATH`.

### Setup

Go dependencies can be added with `GO111MODULE=on go mod vendor`. The Go dependencies will be downloaded to the gitignored `vendor` directory.

Cue dependencies should be added inside of the `/pkg/version/cue.mod` directory by running `cue get go k8s.io/api/...` inside of that directory. The Cue dependencies will be downloaded to the gitignored `/pkg/version/cue.mod/gen` directory.

### Local Quickstart (K3d)

This section outlines how to set up a local development environment in [K3d](https://k3d.io). These instructions are easy enough to replicate for any remote Kubernetes cluster you have cluster administrator access to.

First, run the following commands to set up a local cluster and install necessary resources in it:

```
k3d cluster create gm-operator -a 1 -p 30000:10808@loadbalancer
export KUBECONFIG=$(k3d kubeconfig write gm-operator)
```

Next, ensure you have the following environment variables sourced: `NEXUS_USER` and `NEXUS_PASSWORD`. These require your credentials for pulling Grey Matter core service Docker images from `docker.greymatter.io`.

Run the following to install the operator in your K3d cluster, optionally building and importing a development image from the current code branch.

```
./scripts/dev deploy k3d
```

A sample Mesh custom resource has been provided in `hack/sample-k8s.yaml` for declaring the state of mesh components. Run the following to apply it and view the operator logs for mesh components being created and configured:

```
./scripts/dev sample k3d
```

To uninstall all components and tear down the local cluster, run:

```
./scripts/dev cleanup k3d
k3d cluster delete gm-operator
```

### OpenShift Quickstart

This section documents commands provided for deploying to an [OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift/container-platform) cluster.

First, create a `gm-operator` project in your OpenShift cluster.

```
oc create project gm-operator
```

Next, ensure you have the following environment variables sourced: `NEXUS_USER` and `NEXUS_PASSWORD`. These require your credentials for pulling Grey Matter core service Docker images from `docker.greymatter.io`.

Run the following to install the operator in your OpenShift cluster, optionally building and pushing a development image from the current code branch to the `docker.greymatter.io` repository.

```
./scripts/dev deploy oc
```

The operator will be installed in a custom catalog in your configured OpenShift cluster to be manged by the Operator Lifecycle Manager.

A sample Mesh custom resource has been provided in `hack/sample-openshift.yaml` for declaring the state of mesh components. Run the following to apply it and view the operator logs for mesh components being created and configured:

```
./scripts/dev sample oc
```

To uninstall all components:

```
./scripts/dev cleanup oc
```
