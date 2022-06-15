# Operator

The Grey Matter Operator enables a bootstrapped mesh deployment using the `greymatter.io/v1.Mesh`
CRD to manage mesh deployments in a Kubernetes cluster.

## Prerequisites

- [kubectl v1.23+](https://kubernetes.io/docs/tasks/tools/)
  - Cluster administrative access
- [CUE CLI](https://cuelang.org/docs/install/)

> NOTE: This project makes use of git submodules for dependency management.

## Getting Started

Make sure you have fetched all necessary dependencies:
```bash
./scripts/bootstrap # checks that you have the latest dependencies for the cue evaluation of manifests.
```

Evaluate the kubernetes manifests using CUE:

Note: You may run the operator locally entirely without GitOps (i.e., using only baked-in, local CUE config) by adding
`-t test=true` to the `cue eval ...`. At the moment, you will still need to create the greymatter-sync-secret as below,
or else remove the references to it in the operator manifests, but it won't be used with `-t test=true`.

```bash
( 
cd pkg/cuemodule/core
cue eval -c ./k8s/outputs --out text -e operator_manifests_yaml | kubectl apply -f -

kubectl create secret docker-registry gm-docker-secret \
  --docker-server=quay.io \
  --docker-username=$GREYMATTER_REGISTRY_USERNAME \s
  --docker-password=$GREYMATTER_REGISTRY_PASSWORD \
  --docker-email=$GREYMATTER_REGISTRY_EMAIL \
  -n gm-operator
  
  # EDIT THIS to reflect your own, or some other SSH private key with access,
  # to the repository you would like the operator to use for GitOps. Note
  # that by default, the operator is going to fetch from 
  # https://github.com/greymatter-io/gitops-core and you would
  # need to edit the operator StatefulSet to change the argument to the
  # operator binary to change the git repository or branch.
  kubectl create secret generic greymatter-sync-secret \
  --from-file=id_ed25519=$HOME/.ssh/id_ed25519 \
  -n gm-operator
)
```

> HINT: Your username and password are your Quay.io credentials authorized to the greymatterio organization.

The operator will be running in a pod in the `gm-operator` namespace, and shortly after installation, the default Mesh
CR described in `pkg/cuemodule/core/inputs.cue` will be automatically deployed.

That is all you need to do to launch the operator. Note that if you have the spire config flag set
(in pkg/cuemodule/core/inputs.cue) then you will need to wait for the operator to insert the server-ca bootstrap certificates
before spire-server and spire-agent can successfully launch.


## Alternative Debug Build

If you would like to attach a remote debugger to your operator container, do the following:
```bash
# Builds and pushes quay.io/greymatterio/gm-operator:debug from Dockerfile.debug. Edit to taste.
# You will need to have your credentials in $GREYMATTER_REGISTRY_USERNAME, $GREYMATTER_REGISTRY_EMAIL, and $GREYMATTER_REGISTRY_PASSWORD
./scripts/build debug_container

# Push the image you just built to Nexus
docker push quay.io/greymatterio/gm-operator:latest-debug

# Launch the operator with the debug build in debug mode.
# Note the two tags (`operator_image` and `debug`) which are the only differences from Getting Started
( 
cd pkg/cuemodule
cue eval -c ./k8s/outputs --out text \
         -t operator_image=quay.io/greymatterio/gm-operator:latest-debug \
         -t debug=true \
         -e operator_manifests_yaml | kubectl apply -f -

kubectl create secret docker-registry gm-docker-secret \
  --docker-server=quay.io \
  --docker-username=$GREYMATTER_REGISTRY_USERNAME \
  --docker-password=$GREYMATTER_REGISTRY_PASSWORD \
  --docker-email=$GREYMATTER_REGISTRY_EMAIL \
  -n gm-operator
)
  
# To connect, first port-forward to 2345 on the operator container in a separate terminal window
kubectl port-forward sts/gm-operator 2345 -n gm-operator

# Now you can connect GoLand or VS Code or just vanilla Delve to localhost:2345 for debugging
# Note that the `:debug` container waits for the debugger to connect before running the operator
```

## Inspecting Manifests

The following commands print out manifests that can be applied to a Kubernetes cluster, for your inspection:

```bash
( 
  cd pkg/cuemodule/core
  cue eval -c ./k8s/outputs --out text -e spire_manifests_yaml
)

# pick which manifests you'd like to inspect

(
  cd pkg/cuemodule/core
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
* You should be able to log in to `quay.io`

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
