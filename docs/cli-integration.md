# CLI Integration Plans

1. [Overview](#overview)
2. [Manifests](#manifests)
   1. [Namespace](#namespace)
   2. [CustomResourceDefinition](#customresourcedefinition)
   3. [ServiceAccount](#serviceaccount)
   4. [Role](#role)
   5. [RoleBinding](#rolebinding)
   6. [ClusterRole](#clusterrole)
   7. [ClusterRoleBinding](#clusterrolebinding)
   8. [Secret](#secret)
   9. [Deployment](#deployment)
3. [Creating the Docker Secret](#creating-the-docker-secret)

## Overview

This document outlines plans to integrate the Operator with the CLI [https://github.com/greymatter-io/cli].

In order to manage versioning with the CLI, this repo should expose functions that deploy the Operator and define Mesh CRs in a Kubernetes cluster. The CLI should simply import and run the functions.

The following CLI commands should each call one function exposed by the package:
1. `greymatter operator init [--namespace]` should create the required resources in a Kubernetes cluster for deploying the Operator. `namespace` should default to `gm-operator`.
2. `greymatter operator remove [--namespace]` should delete the resources in a Kubernetes cluster that were used for deploying the Operator.
3. `greymatter operator install <mesh_id> [--namespace] [--profile] [--version]` should create a Mesh CR in the Kubernetes cluster where the Operator is deployed. The Mesh CR can be specified via stdin, but if none is specified it should open the configured `$EDITOR` with all defaults and/or any values specified in flags.
  - `mesh_id` is the name of the Mesh CR and also the name of the `zone` in single-zone deployments.
  - `namespace` should default to `default`.
  - `profile` should default to `default`. It uses base recommended settings.
  - `version` should default to `latest` if not specified, which is the last stable release of Grey Matter in parity with the imported version of the Operator.
4. `greymatter operator upgrade <mesh_id>` should upgrade a Mesh CR in the Kubernetes cluster. As with `install`, a Mesh CR can be specified via stdin or otherwise the user can edit the existing configuration.
5. `greymatter operator uninstall <mesh_id>` should delete the Mesh CR.

## Manifests

Note that following CLI integration, the `manifests` directory in this project will only serve as reference material and should not be applied to a Kubernetes cluster. The directory should still be maintained to continue to leverage [controller-gen](https://book.kubebuilder.io/reference/controller-gen.html) scaffolding for future enhancements such as [Finalizers](https://book.kubebuilder.io/reference/using-finalizers.html), [Webhooks](https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html), as well as [CustomComponentConfig](https://book.kubebuilder.io/component-config-tutorial/tutorial.html).

The manifests defined below should be applied in the Kubernetes cluster.

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: [namespace]
  labels:
    control-plane: gm-operator
```

### CustomResourceDefinition

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: meshes.greymatter.io
spec:
  group: greymatter.io
  names:
    kind: Mesh
    listKind: MeshList
    plural: meshes
    singular: mesh
  scope: Namespaced
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        description: Mesh is the Schema for the meshes API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: Defines the desired state of a Mesh.
            properties:
              image_pull_secret:
                description: The name of the secret used for pulling Grey Matter service
                  Docker images. If not specified, defaults to "docker.secret".
                type: string
              users:
                description: "A list of JWT users to add to the JWT Security service.
                  For example: - label: CN=greymatter,OU=Engineering,O=Decipher Technology
                  Studios,L=Alexandria,ST=Virginia,C=US \t values:     email: [\"engineering@greymatter.io\"]
                  \    org: [\"www.greymatter.io\"]     privilege: [\"root\"]"
                items:
                  properties:
                    label:
                      type: string
                    values:
                      additionalProperties:
                        items:
                          type: string
                        type: array
                      type: object
                  required:
                  - label
                  - values
                  type: object
                type: array
              version:
                description: Which version of Grey Matter to install. If not specified,
                  the latest version will be installed.
                type: string
            type: object
          status:
            description: Defines the observed state of a Mesh.
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
```

### ServiceAccount

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: gm-operator
  namespace: [namespace]
```

### Role

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: gm-leader-election-role
  namespace: [namespace]
rules:
- apiGroups:
  - ""
  - coordination.k8s.io
  resources:
  - configmaps
  - leases
  verbs:
  - get
  - list
  - watch
  - create
  - update
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - events
  verbs:
  - create
  - patch
```

### RoleBinding

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: gm-leader-election-rolebinding
  namespace: [namespace]
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: gm-leader-election-role
subjects:
- kind: ServiceAccount
  name: gm-operator
  namespace: [namespace]
```

### ClusterRole

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: null
  name: gm-operator-role
rules:
- apiGroups:
  - apps
  resources:
  - deployments
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - configmaps
  - services
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - ""
  resources:
  - pods
  - secrets
  - serviceaccounts
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - extensions
  resources:
  - ingresses
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - greymatter.io
  resources:
  - meshes
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
- apiGroups:
  - greymatter.io
  resources:
  - meshes/finalizers
  verbs:
  - update
- apiGroups:
  - greymatter.io
  resources:
  - meshes/status
  verbs:
  - get
  - patch
  - update
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterrolebindings
  - clusterroles
  verbs:
  - create
  - delete
  - get
  - list
  - patch
  - update
  - watch
```

### ClusterRoleBinding

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: gm-operator-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: gm-operator-role
subjects:
- kind: ServiceAccount
  name: gm-operator
  namespace: [namespace]
```

### Secret

See [Creating the Docker Secret](#creating-the-docker-secret) for more information.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: docker.secret
  namespace: [namespace]
type: kubernetes.io/dockerconfigjson
data:
  .dockerconfigjson: [base64_encoded_string]
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: gm-operator
  name: gm-operator
  namespace: [namespace]
spec:
  selector:
    matchLabels:
      control-plane: gm-operator
  replicas: 1
  template:
    metadata:
      labels:
        control-plane: gm-operator
    spec:
      containers:
      - name: gm-operator
        image: docker.greymatter.io/internal/gm-operator:latest
        imagePullPolicy: IfNotPresent
        command:
        - /gm-operator
        args:
        - --leader-elect
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8081
          initialDelaySeconds: 15
          periodSeconds: 20
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8081
            scheme: HTTP
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          limits:
            cpu: 100m
            memory: 30Mi
          requests:
            cpu: 100m
            memory: 20Mi
        securityContext:
          allowPrivilegeEscalation: false
      imagePullSecrets:
      - name: docker.secret
      securityContext:
        runAsNonRoot: true
      serviceAccountName: gm-operator
      terminationGracePeriodSeconds: 10
```

## Creating the Docker Secret

The CLI should read from some predetermined environment variables such as `GREYMATTER_OPERATOR_IMAGE_REGISTRY_USER` and `GREYMATTER_OPERATOR_IMAGE_REGISTRY_PASSWORD`. The CLI config file can also define them via `operator.image_registry.user` and `operator.image_registry.password`. If those aren't found in the environment, it should prompt the user for those values.

From there, the CLI should pass those values into a function defined in this project that encodes the following object into the base64-encoded string:

```json
{
  "auths": {
    "docker.greymatter.io": {
      "username": "$GREYMATTER_OPERATOR_IMAGE_REGISTRY_USER",
      "password": "$GREYMATTER_OPERATOR_IMAGE_REGISTRY_PASSWORD",
      "email": "$GREYMATTER_OPERATOR_IMAGE_REGISTRY_USER"
    }
  }
}
```
