# GM Operator

## Resources

- [Istio's operator spec](https://github.com/istio/api/blob/master/operator/v1alpha1/operator.pb.go#L97)
- [Example controller code for managing a CRD](https://github.com/kubernetes/sample-controller)
- [Tutorial on CRD code gen](https://www.openshift.com/blog/kubernetes-deep-dive-code-generation-customresources)
- [Another tutorial](https://itnext.io/how-to-generate-client-codes-for-kubernetes-custom-resource-definitions-crd-b4b9907769ba)

## API Client Generation

### Setup

Run the following to clone the `kubernetes/code-generator` repo into your $GOPATH and checkout to the latest stable version for building and running the code-generation binaries. As a convenience, the script returns you to this repo's workspace.

```bash
cd $GOPATH/src/github.com
mkdir -p kubernetes
cd kubernetes
git clone git@github.com:kubernetes/code-generator
cd code-generator
cd $GOPATH/src/github.com/bcmendoza/gm-operator
```

### Generate

Finally, **in this workspace** (required so that the code-generation binaries don't fetch this repo from remote), run the script to generate the clientset, informers, and listers for the API defined in `apis/install/v1`:

```bash
./generate.sh
```

Any time changes are made to `/apis/install/v1`, run this script.
