# Grey Matter Operator

Grey Matter Operator is a Kubernetes operator that enables using `greymatter.io/v1alpha1.Mesh` custom resource objects to manage Grey Matter Mesh deployments in a Kubernetes cluster.

This project is currently in an unstable alpha stage. This README will be updated constantly throughout the initial development process. For now, most things documented here will be in short-form.

## Development

### Dependencies

Grey Matter Operator is built using the [Operator SDK](https://sdk.operatorframework.io). As such, this project adheres to its conventions and recommended best practices.

Download the [Operator SDK CLI v1.12.0](https://sdk.operatorframework.io/docs/installation/) for maintaining the project API, manifests, and general project structure.

It also uses [Cue](https://cuelang.org/docs/install/) for maintaining much of its internal API.

If building for [OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift/container-platform), you'll also need the [OpenShift CLI](https://mirror.openshift.com/pub/openshift-v4/x86_64/clients/ocp/).

### Setup

Go dependencies can be added with `GO111MODULE=on go mod vendor`. The Go dependencies will be downloaded to the gitignored `vendor` directory.

Cue dependencies should be added inside of the `/pkg/version/cue.mod` directory by running `cue get go k8s.io/api/...` inside of that directory. The Cue dependencies will be downloaded to the gitignored `/pkg/version/cue.mod/gen` directory.

### Local Quickstart

This section outlines how to set up a local development environment in [K3d](https://k3d.io) (not a requirement for this project, but an alternative to deploying to an online Kubernetes cluster).

First, run the following commands to set up a local cluster and install necessary resources in it:

```
k3d cluster create gm-operator -a 1 -p 30000:10808@loadbalancer
export KUBECONFIG=$(k3d kubeconfig write gm-operator)
make install
```

Next, create a `gm-operator` namespace in your local cluster and create a secret named `gm-docker-secret` in the `gm-operator` namespace that will be used to pull Grey Matter container images from your organization's private Docker registry:

```
kubectl create namespace gm-operator

kubectl create secret docker-registry gm-docker-secret \
  --docker-server=<your-registry-server> \
  --docker-username=<your-username> \
  --docker-password=<your-password> \
  --docker-email=<your-email> -n gm-operator
```

Finally, the following command will build a binary from code and run it from your host to talk to the local cluster from the outside. You can continue to develop the code and run this command as you work on this project:

```
make run
```

To uninstall and tear down the local cluster, exit the terminal process and run:

```
make uninstall
k3d cluster delete gm-operator
```

### OpenShift Quickstart (Decipher only)

This section documents commands provided for deploying to an [OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift/container-platform) cluster.

It currently only supports the internal development process for Decipher employees who have access to our [Nexus repositories](https://nexus.greymatter.io).

First, create a `gm-operator` project in your OpenShift cluster.

Next, create a secret named `gm-docker-secret` in the `gm-operator` namespace that will be used to pull Grey Matter container images from your organization's private Docker registry:

```
kubectl create secret docker-registry gm-docker-secret \
  --docker-server=docker.greymatter.io \
  --docker-username=<your-email> \
  --docker-password=<your-password> \
  --docker-email=<your-email>
```

Finally, the following commands will build an internal container image, push it to Nexus, and install the operator to a custom catalog in your configured OpenShift cluster to be managed by the Operator Lifecycle Manager:

```
./scripts/dev deploy
```

To remove the operator from your configured OpenShift cluster:

```
./scripts/dev cleanup
oc delete namespace gm-operator
```
