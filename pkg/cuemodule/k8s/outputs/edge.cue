package only

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

let Name = "edge"
edge: [
  appsv1.#Deployment & {
    apiVersion: "apps/v1"
    kind: "Deployment"
    metadata: {
      name: Name
      namespace: mesh.spec.install_namespace
    }
    spec: {
      selector: {
        matchLabels: {"greymatter.io/cluster": Name}
      }
      template: {
        metadata: {
          labels: {"greymatter.io/cluster": Name}
        }
        spec: #spire_permission_requests & {
          containers: [
            #sidecar_container_block & { _Name: Name },
          ]
          volumes: [] + #spire_socket_volumes
          imagePullSecrets: [{name: defaults.image_pull_secret_name}]
          serviceAccountName: Name
        }
      }
    }
  },

  // Used to attach SCCs
  corev1.#ServiceAccount & {
    apiVersion: "v1"
    kind: "ServiceAccount"
    metadata: {
      name:      Name
      namespace: mesh.spec.install_namespace
    }
  },
  corev1.#Service & {
    apiVersion: "v1"
    kind: "Service"
    metadata: {
      name: Name
      namespace: mesh.spec.install_namespace
    }
    spec: {
      selector: "greymatter.io/cluster": Name
      type: "LoadBalancer"
      ports: [{
          name: "ingress",
          port: 10808,
          targetPort: 10808
        }]
    }
  }
]
