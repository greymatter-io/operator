// Manifests for the Redis pod

package only

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

let Name = defaults.redis_cluster_name
redis: [
  appsv1.#StatefulSet & {
    apiVersion: "apps/v1"
    kind: "StatefulSet"
    metadata: {
      name: Name
      namespace: mesh.spec.install_namespace
    }
    spec: {
      serviceName: Name
      selector: {
        matchLabels: {"greymatter.io/cluster": Name}
      }
      template: {
        metadata: {
          labels: {"greymatter.io/cluster": Name}
        }
        spec: {
          containers: [  
            // there are multiple in this ensemble! proxy, control, control-api, and redis

            #sidecar_container_block & {
              _Name: Name
              ports: [{ // redis listens on a non-default sidecar port
                name: "proxy"
                containerPort: defaults.ports.redis_ingress
              }]
            },
            {
              name: "redis"
              image: mesh.spec.images.redis
              command: ["redis-server"]
              args: [
                "--appendonly", "yes",
                "--dir", "/data",
                "--logLevel", "verbose"
              ]
              ports: [{ // HACK this port is exposed so the Service (below) can get to it for easy bootstrap
                name: "redis"
                containerPort: 6379
              }]
              imagePullPolicy: defaults.image_pull_policy,
              volumeMounts: [{
                name: "gm-redis-append-dir-\(mesh.metadata.name)",
                mountPath: "/data"
              }]
            } // redis
          ] // containers

          volumes: #spire_socket_volumes
          imagePullSecrets: [{name: defaults.image_pull_secret_name}]
        }
      }
      volumeClaimTemplates: [
        {
          apiVersion: "v1"
          kind: "PersistentVolumeClaim"
          metadata: name: "gm-redis-append-dir-\(mesh.metadata.name)"
          spec: {
            accessModes: ["ReadWriteOnce"],
            resources: requests: storage: "1Gi"
            volumeMode: "Filesystem"
          }
        }
      ]
    }
  },

  // HACK to avoid static configuration during bootstrap, give things direct access to redis
  // Later, use Redis' sidecar
  corev1.#Service & {
    apiVersion: "v1"
    kind: "Service"
    metadata: {
      name: Name
      namespace: mesh.spec.install_namespace
    }
    spec: {
      selector: "greymatter.io/cluster": Name
      ports: [
        {
          name: "redis",
          port: 6379,
          targetPort: 6379
        },
      ]
    }
  }
]