package base

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

manifests: [...#ManifestGroup] & [
  { _c: [edge] }, 
  { _c: [control, control_api] },
  { _c: [catalog] },
  { _c: [dashboard] },
  { _c: [jwt_security] },
  { _c: [redis] },
  { _c: [prometheus] },
]

#ManifestGroup: {
  _c: [...#Component]
  deployment: appsv1.#Deployment & {
    apiVersion: "apps/v1"
    kind: "Deployment"
    metadata: {
      name: _c[0].name
      namespace: Namespace
    }
    spec: {
      selector: matchLabels: {
        "greymatter.io/cluster": _c[0].name
      }
      template: {
        metadata: labels: {
          "greymatter.io/cluster": _c[0].name
        }
        spec: {
          containers: [
            for _, c in _c {
              {
                name: c.name
                image: c.image
                command: c.command
                args: c.args
                ports: [
                  for k, v in c.ports {
                    {
                      name: k
                      containerPort: v
                    }
                  }
                ]
                envFrom: [ for _, v in c.envFrom { v } ]
                env: [
                  for k, v in c.env {
                    {
                      name: k
                      value: v
                    }
                  }
                ]
                resources: c.resources
                volumeMounts: [
                  for k, v in c.volumeMounts {
                    name: k
                    v
                  }
                ]
              }
            }
          ]
          volumes: [
            for _, c in _c if len(c.volumes) > 0 {
              for k, v in c.volumes {
                {
                  name: k
                  v
                }
              }
            }
          ]
        }
      }
    }
  }
  services: [...corev1.#Service] & [
    for _, c in _c {
      {
        apiVersion: "v1"
        kind: "Service"
        metadata: name: c.name
        spec: {
          selector: "greymatter.io/cluster": c.name
          ports: [
            for k, v in c.ports {
              {
                name: k
                protocol: "TCP"
                port: v
                targetPort: v
              }
            }
          ]
        }
      }
    }
  ]
}

sidecar: {
  container: corev1.#Container & {
    name: "sidecar"
    image: proxy.image
    command: proxy.command
    args: proxy.args
    ports: [
      for k, v in proxy.ports {
        {
          name: k
          containerPort: v
        }
      }
    ]
    envFrom: [ for _, v in proxy.envFrom { v } ]
    env: [
      for k, v in proxy.env {
        {
          name: k
          value: v
        }
      }
    ]
    resources: proxy.resources
    volumeMounts: [
      for k, v in proxy.volumeMounts {
        name: k
        v
      }
    ]
  }
  volumes: [...corev1.#Volume] & [
    for k, v in proxy.volumes {
      {
        name: k
        v
      }
    }
  ]
}
