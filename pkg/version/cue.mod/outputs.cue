package base

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

manifests: [...#ManifestGroup] & [
  { _c: [edge] }, 
  { _c: [redis] },
  { _c: [jwt_security] },
  { _c: [control, control_api] },
  { _c: [catalog] },
  { _c: [dashboard] },
  // { _c: [prometheus] }, // TODO
]

#ManifestGroup: {
  // _c and _tmpl are "inputs" following the "function" pattern.
  // See https://cuetorials.com/patterns/functions/.
  _c: [...#Component]
  _tmpl: {
    metadata: {
      name: _c[0].name
      namespace: InstallNamespace
    }
    spec: {
      selector: matchLabels: {
        "greymatter.io/cluster": _c[0].name
      }
      template: {
        metadata: {
          namespace: InstallNamespace
          labels: {
            "greymatter.io/cluster": _c[0].name
          }
        }
        spec: {
          if _c[0].name != "gm-redis" && _c[0].name != "gm-prometheus" {
            imagePullSecrets: [
              { name: ImagePullSecretName }
            ]
          }
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
                env: [
                  for k, v in c.env {
                    {
                      name: k
                      value: v
                    }
                  }
                  for k, v in c.envFrom {
                    {
                      name: k
                      valueFrom: v
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
          if _c[0].name == "control" {
            serviceAccountName: "gm-control"
          }
        }
      }
    }
  }
  if !_c[0].isStatefulset {
    deployment: appsv1.#Deployment & _tmpl & {
      apiVersion: "apps/v1"
      kind: "Deployment"
    }
  }
  if _c[0].isStatefulset {
    statefulset: appsv1.#StatefulSet & _tmpl & {
      apiVersion: "apps/v1"
      kind: "StatefulSet"
      spec: {
        serviceName: _c[0].name
        volumeClaimTemplates: [
          for _, c in _c if len(c.persistentVolumeClaims) > 0 {
            for k, v in c.persistentVolumeClaims {
              {
                apiVersion: "v1"
                kind: "PersistentVolumeClaim"
                metadata: name: k
                spec: v
              }
            }
          }
        ]
      }
    }
  }
  services: [...corev1.#Service] & [
    for _, c in _c {
      {
        apiVersion: "v1"
        kind: "Service"
        metadata: {
          name: c.name
          namespace: InstallNamespace
        }
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
  configMaps: [...corev1.#ConfigMap] & [
    for _, c in _c if len(c.configMaps) > 0 {
      for k, v in c.configMaps {
        {
          apiVersion: "v1"
          kind: "ConfigMap"
          metadata: {
            name: k
            namespace: InstallNamespace
          }
          data: v
        }
      }
    }
  ]
  secrets: [...corev1.#Secret] & [
    for _, c in _c if len(c.secrets) > 0 {
      for k, v in c.secrets {
        {
          apiVersion: "v1"
          kind: "Secret"
          metadata: {
            name: k
            namespace: InstallNamespace
          }
          stringData: v
        }
      }
    }
  ]
}

sidecar: {
  xdsCluster: string
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
      {
        name: "XDS_CLUSTER"
        value: xdsCluster
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
