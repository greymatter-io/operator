# Operator

The Grey Matter Operator enables a bootstrapped mesh deployment using the `greymatter.io/v1alpha1.Mesh`
CRD to manage mesh deployments in a Kubernetes cluster.

## Prerequisites

- [kubectl v1.23+](https://kubernetes.io/docs/tasks/tools/)
  - Cluster administrative access
- [CUE CLI](https://cuelang.org/docs/install/)

> NOTE: This project makes use of git submodules for dependency management.

## Getting Started

Make sure you have fetched all necessary dependencies:
```bash
./scripts/bootstrap # this makes sure you have the latest dependencies for the cue evaluation of manifests.
```

Create a Grey Matter namespace in your k8s cluster:
```bash
kubectl create namespace greymatter
```

Evaluate the kubernetes manifests using CUE: 
```bash
( 
  cd pkg/cuemodule
  cue eval -c ./k8s/outputs --out text -e operator_manifests_yaml | kubectl apply -f -
)
```

Create the necessary pull secret for the Grey Matter core services:
```bash
kubectl create secret docker-registry gm-docker-secret \
  --docker-server=docker.greymatter.io \
  --docker-username=$GREYMATTER_REGISTRY_USERNAME \
  --docker-password=$GREYMATTER_REGISTRY_PASSWORD \
  --docker-email=$GREYMATTER_REGISTRY_USERNAME \
  -n gm-operator
```
> HINT: Your username and password are your Grey Matter credentials.

The operator will be running in a pod in the `gm-operator` namespace, and shortly after installation, the default Mesh
CR described in `pkg/cuemodule/new_structure/inputs.cue` will be automatically deployed.

## Inspecting Manifests

The following commands print out manifests that can be applied to a Kubernetes cluster:

```bash
( 
  cd pkg/cuemodule
  cue eval -c ./k8s/outputs --out text -e spire_manifests_yaml
)

# pick which manifests you'd like to inspect

(
  cd pkg/cuemodule
  cue eval -c ./k8s/outputs --out text -e operator_manifests_yaml
)
```

OR with Kustomize:

```bash
kubectl kustomize config/context/kubernetes
```
>NOTE: If deploying to OpenShift, you can 
> replace `config/context/kubernetes` with `config/context/openshift`.)

## Using nix-shell

For those using the [Nix package manager](https://nixos.org/download.html), a `shell.nix` script has
been provided at the root of this project to launch the operator in a local
[KinD](https://kind.sigs.k8s.io/) cluster.

Some caveats:
* You should have Docker and Nix installed
* You should be able to login to `docker.greymatter.io`

To launch, simply run:
```bash
nix-shell
```


# Development

## Prerequisites

- [Go v1.17+](https://golang.org/dl/)
- [Operator SDK v1.12+](https://sdk.operatorframework.io)
- Grey Matter CLI 
- [CUE CLI](https://cuelang.org/docs/install/)
- [staticcheck](https://staticcheck.io/)
- git

If building for
[OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift/container-platform),
you'll also need the [OpenShift
CLI](https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/).

## Setup

Verify all dependency installations and update CUE modules:
```
./scripts/bootstrap
```
