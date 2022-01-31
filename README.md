# Grey Matter Operator

Grey Matter Operator is a Kubernetes operator that enables using `greymatter.io/v1alpha1.Mesh`
custom resource objects to manage Grey Matter Mesh deployments in a Kubernetes cluster.

This project is currently in an unstable alpha stage. This README will be updated constantly
throughout the initial development process. For now, most things documented here will be in
short-form.

## Prerequisites

It is assumed that you have [kubectl v1.21+](https://kubernetes.io/docs/tasks/tools/) installed with cluster administrator access.

Next, ensure you have the following environment variables sourced: `GREYMATTER_REGISTRY_USERNAME` and
`GREYMATTER_REGISTRY_PASSWORD`. These require your credentials for pulling Grey Matter core service
Docker images from `docker.greymatter.io`.

## Quick Install

To get the latest stable development version of the operator up and running in your Kubernetes
cluster, run the following:

```
kubectl apply -k config/context/kubernetes

kubectl create secret docker-registry gm-docker-secret \
  --docker-server=docker.greymatter.io \
  --docker-username=$GREYMATTER_REGISTRY_USERNAME \
  --docker-password=$GREYMATTER_REGISTRY_PASSWORD \
  --docker-email=$GREYMATTER_REGISTRY_USERNAME \
  -n gm-operator
```

The operator will be running in a pod in the `gm-operator` namespace.

## Inspecting Manifests

The following command prints the manifests that should be applied to a Kubernetes cluster:

```
kubectl apply -k config/context/kubernetes --dry-run=client -o yaml
```

(NOTE: If deploying to OpenShift, you can replace `config/context/kubernetes` with
`config/context/openshift`.)

Under the hood, kubectl uses [kustomize](https://kustomize.io). As a convenience, `kustomize` may be
downloaded to this repo's `bin` directory by using the `make kustomize` target. To generate the raw
manifests that can be piped into a file after downloading kustomize, run

```
./bin/kustomize build config/context/kubernetes
```

## Development

### Dependencies

Grey Matter Operator is built with [Go 1.17](https://golang.org/dl/).

It has been scaffolded using the [Operator SDK](https://sdk.operatorframework.io). Download the
[Operator SDK CLI v1.12.0](https://sdk.operatorframework.io/docs/installation/) for maintaining the
project API, manifests, and general project structure.

It also uses [Cue](https://cuelang.org/docs/install/) for maintaining much of its internal API.

If building for
[OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift/container-platform),
you'll also need the [OpenShift
CLI](https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/).

### Setup

Go dependencies can be added with `GO111MODULE=on go mod vendor`. The Go dependencies will be
downloaded to the gitignored `vendor` directory.

Cue dependencies should be added inside of the `/pkg/version/cue.mod` directory by running `cue get
go k8s.io/api/...` inside of that directory. The Cue dependencies will be downloaded to the
gitignored `/pkg/version/cue.mod/gen` directory.

### Quickstart (Kubernetes)

This section documents commands provided for deploying to a Kubernetes cluster.

Run the following to install the operator in your Kubernetes cluster, optionally building and
pushing a development image from the current code branch to the `docker.greymatter.io` repository.

```
./scripts/dev deploy k8s
```

The operator will be installed to the `gm-operator` namespace. View its running logs:

```
./scripts/dev logs k8s
```

Run the following to create a Mesh custom resource in your cluster:

```
cat <<EOF | kubectl apply -f -
apiVersion: greymatter.io/v1alpha1
kind: Mesh
metadata:
  name: mesh-sample
spec:
  release_version: '1.7'
  zone: default-zone
  install_namespace: default
EOF
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

From there, all other commands are nearly identical to the Kubernetes quickstart section above, but
with `k3d` instead of `k8s`. You can run the following:

```
./scripts/dev deploy k3d
./scripts/dev logs k3d
./scripts/dev sample k3d
./scripts/dev cleanup k3d
```

Note that the option to build and push a Docker image will simply import your built image into your
local K3d cluster rather than push it to the `docker.greymatter.io` repository.

To tear down the local cluster, run:

```
k3d cluster delete gm-operator
```


### OpenShift Quickstart

This section documents commands provided for deploying to an
[OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift/container-platform)
cluster.

First, create a `gm-operator` project in your OpenShift cluster.

```
oc create project gm-operator
```

From there, all other commands are nearly identical to the Kubernetes quickstart section above, but
with `oc` instead of `k8s`. You can run the following:

```
./scripts/dev deploy oc
./scripts/dev logs oc
./scripts/dev sample oc
./scripts/dev cleanup oc
```

The operator will be installed in a custom catalog in your configured OpenShift cluster to be manged
by the Operator Lifecycle Manager.
