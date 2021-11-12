# Grey Matter Operator

Grey Matter Operator is a Kubernetes operator that enables using `greymatter.io/v1alpha1.Mesh` custom resource objects to manage Grey Matter Mesh deployments in a Kubernetes cluster.

This project is currently in an unstable alpha stage. This README will be updated constantly throughout the initial development process. For now, most things documented here will be in short-form.

## Generating Manifests

The current process for generating manifests to be applied to a Kubernetes cluster for installing the operator is to use [kustomize](https://kustomize.io/). The following command prints the manifests to stdout:

```
kustomize build config/k8s
```

(NOTE: If deploying to OpenShift, you can run `kustomize build config/openshift`.)

As a convenience, `kustomize` may be downloaded to this repo's `bin` directory by using the `make kustomize` target.

When applying these manifests to your Kubernetes cluster, you'll also need to create a [docker-registry secret](https://kubernetes.io/docs/concepts/configuration/secret/#docker-config-secrets) so the Kubelet has access to pull Grey Matter images. Provided that you have the necessary Docker username and password, these can be created with the following command:

```
kubectl create secret docker-registry gm-docker-secret \
  --docker-server=docker.greymatter.io \
  --docker-username=<user> \
  --docker-password=<password> \
  --docker-email=<user> \
  -n gm-operator
```

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

### Quickstart (Kubernetes)

This section documents commands provided for deploying to a Kubernetes cluster.

First, ensure you have the following environment variables sourced: `NEXUS_USER` and `NEXUS_PASSWORD`. These require your credentials for pulling Grey Matter core service Docker images from `docker.greymatter.io`.

Run the following to install the operator in your Kubernetes cluster, optionally building and pushing a development image from the current code branch to the `docker.greymatter.io` repository.

```
./scripts/dev deploy k8s
```

The operator will be installed to the `gm-operator` namespace. View its running logs:

```
./scripts/dev logs k8s
```

A sample Mesh custom resource has been provided in `hack/sample-k8s.yaml` for declaring the state of a mesh's components. Run the following to apply it and view the operator logs for mesh components being created and configured:

```
./scripts/dev sample k8s
```

To uninstall all components:

```
./scripts/dev cleanup k8s
```


### Local Quickstart (K3d)

This section outlines how to set up a local development environment in [K3d](https://k3d.io).

First, run the following commands to set up a local cluster and install necessary resources in it:

```
k3d cluster create gm-operator -a 1 -p 30000:10808@loadbalancer
export KUBECONFIG=$(k3d kubeconfig write gm-operator)
```

From there, all other commands are nearly identical to the Kubernetes quickstart section above, but with `k3d` instead of `k8s`. You can run the following:

```
./scripts/dev deploy k3d
./scripts/dev logs k3d
./scripts/dev sample k3d
./scripts/dev cleanup k3d
```

Note that the option to build and push a Docker image will simply import your built image into your local K3d cluster rather than push it to the `docker.greymatter.io` repository.

To tear down the local cluster, run:

```
k3d cluster delete gm-operator
```


### OpenShift Quickstart

This section documents commands provided for deploying to an [OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift/container-platform) cluster.

First, create a `gm-operator` project in your OpenShift cluster.

```
oc create project gm-operator
```

From there, all other commands are nearly identical to the Kubernetes quickstart section above, but with `oc` instead of `k8s`. You can run the following:

```
./scripts/dev deploy oc
./scripts/dev logs oc
./scripts/dev sample oc
./scripts/dev cleanup oc
```

The operator will be installed in a custom catalog in your configured OpenShift cluster to be manged by the Operator Lifecycle Manager.
