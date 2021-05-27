# GM Operator

## MVP

The goal for this Frackday will be to use the Operator Framework to create an Kubernetes operator that watches for install.greymatter.io/v1.Mesh CR (Custom Resource) objects in a cluster and installs a *subset* of core Grey Matter services into the same namespace where the Mesh CR exists.

By *subset*, we mean that the Operator will deploy the following per Mesh CR:
1. Control
2. Control API with Sidecar
3. Edge

Furthermore, the Operator will spawn a process that will call the Control API to configure the mesh network from Edge to Control API with the goal that we'll be able to access Control API via `{edge}/services/control-api/latest`.

## Resources

- [Operator Framework: Go Operator Tutorial](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/)
- [Operator SDK Installation](https://sdk.operatorframework.io/docs/building-operators/golang/installation/)
- [Operator Manager Overview](https://book.kubebuilder.io/cronjob-tutorial/empty-main.html)
- [Istio's operator spec](https://github.com/istio/api/blob/master/operator/v1alpha1/operator.pb.go#L97)
