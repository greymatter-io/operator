package base

import (
	appsv1        "k8s.io/api/apps/v1"
	corev1        "k8s.io/api/core/v1"
	networkingv1  "k8s.io/api/networking/v1"
)

manifests: [...#ManifestGroup] & [
  { _c: [edge] }, 
  { _c: [redis] },
  { _c: [jwt_security] },
  // Add Control and Control API into the same Deployment.
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
      annotations: {
        for k, v in _c[0].annotations if len(_c[0].annotations) > 0 {
          "\(k)": v
        }
      }
      labels: {
        "app.kubernetes.io/name": _c[0].name
        "app.kubernetes.io/part-of": "greymatter"
        // TODO: Tag with version prior to first release.
        "app.kubernetes.io/created-by": "gm-operator"
        "app.kubernetes.io/managed-by": "gm-operator"
      }
    }
    spec: {
      selector: matchLabels: {
        "greymatter.io/component": _c[0].name
      }
      template: {
        metadata: {
          namespace: InstallNamespace
          labels: {
            "greymatter.io/component": _c[0].name
          }
        }
        spec: {
          if _c[0].name != "gm-redis" && _c[0].name != "gm-prometheus" {
            imagePullSecrets: [
              { name: "gm-docker-secret" }
            ]
          }
          if Environment != "openshift" && (_c[0].name == "gm-redis" || _c[0].name == "gm-prometheus") {
            securityContext: {
              fsGroup: 2000
            }
          }
          containers: [
            for _, c in _c {
              {
                name: c.name
                image: c.image
                if c.command != "" {
                  command: [c.command]
                }
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
  service: corev1.#Service & {
    _name: _c[0].name
    apiVersion: "v1"
    kind: "Service"
    metadata: {
      name: _name
      namespace: InstallNamespace
    }
    spec: {
      selector: "greymatter.io/component": _name
      // Make the edge service a LoadBalancer for ingress
      if _name == "edge" {
        type: "LoadBalancer"
      }
      ports: [
        {
          name: "proxy"
          protocol: "TCP"
          port: 10808
          targetPort: 10808
        },

        for _, c in _c if _c[0].name == "control" || _c[0].name == "catalog" {
          for k, v in c.ports {
            {
              name: k
              protocol: "TCP"
              port: v
              targetPort: v
            },
          }
        }

        if _name == "gm-redis" {
          {
            name: "bootstrap"
            protocol: "TCP"
            port: 10707
            targetPort: 10707
          },
        }

        // if _name == "control" {
        //   {
        //     name: "api"
        //     protocol: "TCP"
        //     port: 5555
        //     targetPort: 5555
        //   },
        // }
        // if _name == "control" {
        //   {
        //     name: "xds"
        //     protocol: "TCP"
        //     port: 50000
        //     targetPort: 50000
        //   },
        // }
        // if _name == "catalog" {
        //   {
        //     name: "api"
        //     protocol: "TCP"
        //     port: 8080
        //     targetPort: 8080
        //   },
        // }
        // if _name == "gm-redis" {
        //   {
        //     name: "gm-redis"
        //     protocol: "TCP"
        //     port: 6379
        //     targetPort: 6379
        //   },
        // }
      ]
    }
  }
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
  if _c[0].name == "edge" {
    ingress: networkingv1.#Ingress & {
      {
        apiVersion: "networking.k8s.io/v1"
        kind: "Ingress"
        metadata: {
          name: MeshName
          namespace: InstallNamespace
        }
        spec: {
          rules: [
            {
              host: IngressSubDomain
              http: paths: [
                {
                  pathType: "ImplementationSpecific"
                  backend: {
                    service: {
                      name: "edge"
                      port: number: 10808
                    }
                  }
                }
              ]
            }
          ]
        }
      }
    }
  }
}

sidecar: {
  xdsCluster: string
  // localPort: *10808 | int32
  node: *"" | string
  controlHost: *"control.\(InstallNamespace).svc.cluster.local" | string
  if xdsCluster == "control" {
    staticConfig: envoyMeshConfig
    localPort: 5555
  }
  if xdsCluster == "catalog" {
    staticConfig: envoyMeshConfig
    localPort: 8080
  }
  if xdsCluster == "gm-redis" {
    staticConfig: envoyRedis
  }
  container: corev1.#Container & {
    name: "sidecar"
    image: proxy.image
    if proxy.command != "" {
      command: [proxy.command]
    }
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
