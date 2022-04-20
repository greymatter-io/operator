// K8s manifests for Spire server and agent, permissions and volumes

package only

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
  admisv1 "k8s.io/api/admissionregistration/v1"
  // "encoding/base64"
)

operator_namespace: [
  corev1.#Namespace & {
    apiVersion: "v1"
    kind:       "Namespace"
    metadata: {
      labels: name: "gm-operator"
      name: "gm-operator"
    }
  }
]

operator_crd: [
  // TODO we need the #CustomResourceDefinition definition
  {
    apiVersion: "apiextensions.k8s.io/v1"
    kind:       "CustomResourceDefinition"
    metadata: {
      annotations: "controller-gen.kubebuilder.io/version": "v0.6.1"
      name: "meshes.greymatter.io"
    }
    spec: {
      conversion: {
        strategy: "Webhook"
        webhook: {
          clientConfig: service: {
            name:      "gm-webhook"
            namespace: "gm-operator"
            path:      "/convert"
          }
          conversionReviewVersions: [
            "v1",
          ]
        }
      }
      group: "greymatter.io"
      names: {
        kind:     "Mesh"
        listKind: "MeshList"
        plural:   "meshes"
        singular: "mesh"
      }
      scope: "Cluster"
      versions: [{
        additionalPrinterColumns: [{
          jsonPath: ".spec.install_namespace"
          name:     "Install Namespace"
          type:     "string"
        }, {
          jsonPath: ".spec.release_version"
          name:     "Release Version"
          type:     "string"
        }, {
          jsonPath: ".spec.zone"
          name:     "Zone"
          type:     "string"
        }]
        name: "v1alpha1"
        schema: openAPIV3Schema: {
          description: "Mesh defines a Grey Matter mesh's desired state and describes its observed state."

          properties: {
            apiVersion: {
              description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources"

              type: "string"
            }
            kind: {
              description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"

              type: "string"
            }
            metadata: type: "object"
            spec: {
              description: "MeshSpec defines the desired state of a Grey Matter mesh."
              properties: {
                image_pull_secrets: {
                  description: "A list of pull secrets to try for fetching core services."
                  items: type: "string"
                  type: "array"
                }
                images: {
                  description: "A list of OCI image strings and their respective pull secret names. These are treated as overrides to the specified \"release_version\"."

                  properties: {
                    catalog: type: "string"
                    control: type: "string"
                    control_api: type: "string"
                    dashboard: type: "string"
                    jwt_security: type: "string"
                    prometheus: type: "string"
                    proxy: type: "string"
                    redis: type: "string"
                  }
                  type: "object"
                }
                install_namespace: {
                  description: "Namespace where mesh core components and dependencies should be installed."

                  type: "string"
                }
                release_version: {
                  default:     "latest"
                  description: "The version of Grey Matter to install for this mesh."
                  enum: [
                    "1.6",
                    "1.7",
                    "latest",
                  ]
                  type: "string"
                }
                user_tokens: {
                  description: "Add user tokens to the JWT Security Service."
                  items: {
                    properties: {
                      label: type: "string"
                      values: {
                        additionalProperties: {
                          items: type: "string"
                          type: "array"
                        }
                        type: "object"
                      }
                    }
                    required: [
                      "label",
                      "values",
                    ]
                    type: "object"
                  }
                  type: "array"
                }
                watch_namespaces: {
                  description: "Namespaces to include in the mesh network."
                  items: type: "string"
                  type: "array"
                }
                zone: {
                  default:     "default-zone"
                  description: "Label this mesh as belonging to a particular zone."
                  type:        "string"
                }
              }
              required: [
                "install_namespace",
                "release_version",
                "zone",
              ]
              type: "object"
            }
            status: {
              description: "MeshStatus describes the observed state of a Grey Matter mesh."

              type: "object"
            }
          }
          type: "object"
        }
        served:  true
        storage: true
        subresources: status: {}
      }]
    }
    status: {
      acceptedNames: {
        kind:   ""
        plural: ""
      }
      conditions: []
      storedVersions: []
    }
  },
]

operator_k8s: [
  appsv1.#StatefulSet & {
    apiVersion: "apps/v1"
    kind:       "StatefulSet"
    metadata: {
      labels: name: "gm-operator"
      name:      "gm-operator"
      namespace: "gm-operator"
    }
    spec: {
      serviceName: "gm-operator"
      replicas:    1
      selector: matchLabels: name: "gm-operator"
      template: {
        metadata: labels: name: "gm-operator"
        spec: {
          containers: [{
            command: [
              "/app/operator",
            ]
            image: defaults.images.operator
            imagePullPolicy: "IfNotPresent"
            livenessProbe: {
              httpGet: {
                path: "/healthz"
                port: 8081
              }
              initialDelaySeconds: 120
              periodSeconds:       20
            }
            name: "operator"
            ports: [{
              containerPort: 9443
              name:          "webhook-server"
              protocol:      "TCP"
            }]
            readinessProbe: {
              httpGet: {
                path: "/readyz"
                port: 8081
              }
              initialDelaySeconds: 120
              periodSeconds:       10
            }
            resources: {
              limits: {
                cpu:    "200m"
                memory: "300Mi"
              }
              requests: {
                cpu:    "100m"
                memory: "150Mi"
              }
            }
            securityContext: allowPrivilegeEscalation: false
            volumeMounts: [{
              mountPath: "/tmp/k8s-webhook-server/serving-certs"
              name:      "webhook-cert"
              readOnly:  true
            }]
          }]
          imagePullSecrets: []
          securityContext: runAsNonRoot: true
          serviceAccountName:            "gm-operator"
          terminationGracePeriodSeconds: 10
          volumes: [{
            name: "webhook-cert"
            secret: {
              defaultMode: 420
              items: [{
                key:  "tls.crt"
                path: "tls.crt"
              }, {
                key:  "tls.key"
                path: "tls.key"
              }]
              secretName: "gm-webhook-cert"
            }
          }]
        }
      }
    }
  },
  corev1.#ServiceAccount & {
    apiVersion: "v1"
    imagePullSecrets: [{
      name: "gm-docker-secret"
    }]
    kind: "ServiceAccount"
    metadata: {
      name:      "gm-operator"
      namespace: "gm-operator"
    }
  },
  rbacv1.#Role & {
    apiVersion: "rbac.authorization.k8s.io/v1"
    kind:       "Role"
    metadata: {
      name:      "gm-leader-election-role"
      namespace: "gm-operator"
    }
    rules: [{
      apiGroups: [
        "",
      ]
      resources: [
        "configmaps",
      ]
      verbs: [
        "get",
        "list",
        "watch",
        "create",
        "update",
        "patch",
        "delete",
      ]
    }, {
      apiGroups: [
        "coordination.k8s.io",
      ]
      resources: [
        "leases",
      ]
      verbs: [
        "get",
        "list",
        "watch",
        "create",
        "update",
        "patch",
        "delete",
      ]
    }, {
      apiGroups: [
        "",
      ]
      resources: [
        "events",
      ]
      verbs: [
        "create",
        "patch",
      ]
    }]
  },
  rbacv1.#ClusterRole & {
    apiVersion: "rbac.authorization.k8s.io/v1"
    kind:       "ClusterRole"
    metadata: name: "gm-operator-role"
    rules: [{
      apiGroups: [
        "apiextensions.k8s.io",
      ]
      resourceNames: [
        "meshes.greymatter.io",
      ]
      resources: [
        "customresourcedefinitions",
      ]
      verbs: [
        "get",
      ]
    }, {
      apiGroups: [
        "greymatter.io",
      ]
      resources: [
        "meshes",
      ]
      verbs: [
        "create",
        "delete",
        "get",
        "list",
        "patch",
        "update",
        "watch",
      ]
    }, {
      apiGroups: [
        "greymatter.io",
      ]
      resources: [
        "meshes/status",
      ]
      verbs: [
        "get",
        "patch",
        "update",
      ]
    }, {
      apiGroups: [
        "admissionregistration.k8s.io",
      ]
      resourceNames: [
        "gm-mutate-config",
        "gm-validate-config",
      ]
      resources: [
        "mutatingwebhookconfigurations",
        "validatingwebhookconfigurations",
      ]
      verbs: [
        "get",
        "patch",
      ]
    }, {
      apiGroups: [
        "apps",
      ]
      resources: [
        "deployments",
        "statefulsets",
      ]
      verbs: [
        "get",
        "list",
        "create",
        "update",
      ]
    }, {
      apiGroups: [
        "",
      ]
      resources: [
        "configmaps",
        "secrets",
        "serviceaccounts",
        "services",
      ]
      verbs: [
        "get",
        "create",
        "update",
        "patch",
      ]
    }, {
      apiGroups: [
        "rbac.authorization.k8s.io",
      ]
      resources: [
        "clusterrolebindings",
        "clusterroles",
      ]
      verbs: [
        "get",
        "create",
        "update",
      ]
    }, {
      apiGroups: [
        "",
      ]
      resources: [
        "pods",
      ]
      verbs: [
        "list",
      ]
    }, {
      apiGroups: [
        "networking.k8s.io",
      ]
      resources: [
        "ingresses",
      ]
      verbs: [
        "get",
        "create",
        "update",
      ]
    }, {
      apiGroups: [
        "config.openshift.io",
      ]
      resources: [
        "ingresses",
      ]
      verbs: [
        "list",
      ]
    }, {
      apiGroups: [
        "",
      ]
      resources: [
        "namespaces",
      ]
      verbs: [
        "get",
        "create",
      ]
    }, {
      apiGroups: [
        "apps",
      ]
      resources: [
        "daemonsets",
      ]
      verbs: [
        "get",
        "create",
      ]
    }, {
      apiGroups: [
        "rbac.authorization.k8s.io",
      ]
      resources: [
        "roles",
        "rolebindings",
      ]
      verbs: [
        "get",
        "create",
      ]
    }, {
      apiGroups: [
        "",
      ]
      resources: [
        "configmaps",
      ]
      verbs: [
        "list",
      ]
    }, {
      apiGroups: [
        "authentication.k8s.io",
      ]
      resources: [
        "tokenreviews",
      ]
      verbs: [
        "get",
        "create",
      ]
    }, {
      apiGroups: [
        "",
      ]
      resources: [
        "nodes",
        "nodes/proxy",
        "pods",
      ]
      verbs: [
        "get",
        "list",
        "watch",
      ]
    }]
  },
  rbacv1.#RoleBinding & {
    apiVersion: "rbac.authorization.k8s.io/v1"
    kind:       "RoleBinding"
    metadata: {
      name:      "gm-leader-election-rolebinding"
      namespace: "gm-operator"
    }
    roleRef: {
      apiGroup: "rbac.authorization.k8s.io"
      kind:     "Role"
      name:     "gm-leader-election-role"
    }
    subjects: [{
      kind:      "ServiceAccount"
      name:      "gm-operator"
      namespace: "gm-operator"
    }]
  },
  rbacv1.#ClusterRoleBinding & {
    apiVersion: "rbac.authorization.k8s.io/v1"
    kind:       "ClusterRoleBinding"
    metadata: name: "gm-operator-rolebinding"
    roleRef: {
      apiGroup: "rbac.authorization.k8s.io"
      kind:     "ClusterRole"
      name:     "gm-operator-role"
    }
    subjects: [{
      kind:      "ServiceAccount"
      name:      "gm-operator"
      namespace: "gm-operator"
    }]
  },
  corev1.#Secret & { // the values here get filled in programmatically by the operator
    apiVersion: "v1"
    data: {
      "tls.crt": ''
      "tls.key": ''
    }
    kind: "Secret"
    metadata: {
      name:      "gm-webhook-cert"
      namespace: "gm-operator"
    }
  },
  corev1.#Service & {
    apiVersion: "v1"
    kind:       "Service"
    metadata: {
      name:      "gm-webhook"
      namespace: "gm-operator"
    }
    spec: {
      ports: [{
        port:       443
        protocol:   "TCP"
        targetPort: 9443
      }]
      selector: name: "gm-operator"
    }
  },
  admisv1.#MutatingWebhookConfiguration & {
    apiVersion: "admissionregistration.k8s.io/v1"
    kind:       "MutatingWebhookConfiguration"
    metadata: name: "gm-mutate-config"
    webhooks: [{
      admissionReviewVersions: [
        "v1",
        "v1beta1",
      ]
      clientConfig: service: {
        name:      "gm-webhook"
        namespace: "gm-operator"
        path:      "/mutate-workload"
      }
      failurePolicy: "Fail"
      name:          "mutate-workload.greymatter.io"
      namespaceSelector: matchExpressions: [{
        key:      "name"
        operator: "NotIn"
        values: [
          "gm-operator",
          "spire",
        ]
      }]
      rules: [{
        apiGroups: [
          "",
          "apps",
        ]
        apiVersions: [
          "v1",
        ]
        operations: [
          "CREATE",
          "UPDATE",
          "DELETE",
        ]
        resources: [
          "pods",
          "deployments",
          "statefulsets",
        ]
      }]
      sideEffects: "None"
    }, {
      admissionReviewVersions: [
        "v1",
        "v1beta1",
      ]
      clientConfig: service: {
        name:      "gm-webhook"
        namespace: "gm-operator"
        path:      "/mutate-mesh"
      }
      failurePolicy: "Fail"
      name:          "mutate-mesh.greymatter.io"
      rules: [{
        apiGroups: [
          "greymatter.io",
        ]
        apiVersions: [
          "v1alpha1",
        ]
        operations: [
          "CREATE",
          "UPDATE",
        ]
        resources: [
          "meshes",
        ]
      }]
      sideEffects: "None"
    }]
  },
  
  admisv1.#ValidatingWebhookConfiguration & {
    apiVersion: "admissionregistration.k8s.io/v1"
    kind:       "ValidatingWebhookConfiguration"
    metadata: name: "gm-validate-config"
    webhooks: [{
      admissionReviewVersions: [
        "v1",
        "v1beta1",
      ]
      clientConfig: service: {
        name:      "gm-webhook"
        namespace: "gm-operator"
        path:      "/validate-mesh"
      }
      failurePolicy: "Fail"
      name:          "validate-mesh.greymatter.io"
      rules: [{
        apiGroups: [
          "greymatter.io",
        ]
        apiVersions: [
          "v1alpha1",
        ]
        operations: [
          "CREATE",
          "UPDATE",
          "DELETE",
        ]
        resources: [
          "meshes",
        ]
      }]
      sideEffects: "None"
    }]
  },
]
