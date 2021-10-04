package base

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

sidecar: corev1.#Container & {
  name: "sidecar"
}

manifests: [...#ManifestGroup] & [
  { _components: [edge] }, 
  { _components: [control, control_api] },
  { _components: [catalog] },
  { _components: [dashboard] },
  { _components: [jwt_security] },
  { _components: [redis] },
  { _components: [prometheus] },
]

#ManifestGroup: {
  _components: [...#Component]
  deployment: appsv1.#Deployment & {
    apiVersion: "apps/v1"
    kind: "Deployment"
    metadata: name: _components[0].name
    spec: {
      selector: matchLabels: {
        "greymatter.io/component": _components[0].name
        "greymatter.io/cluster": _components[0].name
      }
      template: {
        metadata: labels: {
          "greymatter.io/component": _components[0].name
          "greymatter.io/cluster": _components[0].name
        }
        spec: {
          containers: [
            for _, c in _components {
              {
                name: c.name
                image: c.image
                command: c.command
                args: c.args
                ports: [
                  for k, v in c.ports {
                    { name: k, containerPort: v }
                  }
                ]
                envFrom: [ for _, v in c.envFrom { v } ]
                env: [
                  for k, v in c.env {
                    { name: k, value: v }
                  }
                ]
                // resources: c.resources
                // volumeMounts: c.volumeMounts
              }
            }
          ]
        }
      }
    }
  }
  // services: [...corev1.#Service] & [
  //   for _, c in _components {
  //     {
  //       apiVersion: "v1"
  //       kind: "Service"
  //       metadata: name: c.name
  //       spec: {
  //         selector: "greymatter.io/component": c.name
  //         ports: [
  //           for name, port in c.ports {
  //             {
  //               name: name
  //               protocol: "TCP"
  //               port: port
  //               targetPort: port
  //             }
  //           }
  //         ]
  //       }
  //     }
  //   }
  // ]
}
