package base

import (
	corev1 "k8s.io/api/core/v1"
)

#Component: {
  name: string
  isStatefulset: bool
  image: string
  command: [...string]
  args: [...string]
  resources: corev1.#Resources
  ports: [string]: int32
  envFrom: [string]: corev1.#EnvVarSource
  env: [string]: string
  volumeMounts: [string]: corev1.#VolumeMount
  volumes: [string]: corev1.#VolumeSource
  persistentVolumeClaims: [string]: corev1.#PersistentVolumeClaimSpec
  configMaps: [string]: [string]: string
  secrets: [string]: [string]: string
}

proxy: #Component & {
  image: =~"^docker.greymatter.io/(release|development)/gm-proxy:" & !~"latest$"
  isStatefulset: false
  ports: proxy: MeshPort
  env: {
    ENVOY_ADMIN_LOG_PATH: "/dev/stdout",
    PROXY_DYNAMIC: "true"
    XDS_ZONE: Zone
    XDS_HOST: "control.\(InstallNamespace).svc.cluster.local"
    XDS_PORT: "50000"
  }
  if Spire {
    env: SPIRE_PATH: "/run/spire/socket/agent.sock"
    volumeMounts: "spire-socket": mountPath: "/run/spire/socket"
    volumes: "spire-socket": hostPath: {
      path: "/run/spire/socket"
      type: "DirectoryOrCreate"
    }
  }
}

edge: #Component & proxy & {
  name: "edge"
  env: XDS_CLUSTER: "edge"
}

control: #Component & {
  name: "control"
  isStatefulset: false
  image: =~"^docker.greymatter.io/(release|development)/gm-control:" & !~"latest$"
  ports: xds: 50000
  env: {
    GM_CONTROL_CMD: "kubernetes"
    GM_CONTROL_KUBERNETES_NAMESPACES: WatchNamespaces
    GM_CONTROL_KUBERNETES_CLUSTER_LABEL: "greymatter.io/cluster"
    GM_CONTROL_KUBERNETES_PORT_NAME: "proxy"
    GM_CONTROL_XDS_ADS_ENABLED: "true"
    GM_CONTROL_XDS_RESOLVE_DNS: "true"
    GM_CONTROL_API_HOST: "127.0.0.1:5555"
    GM_CONTROL_API_INSECURE: "true"
    GM_CONTROL_API_SSL: "false"
    GM_CONTROL_API_KEY: "xxx"
    GM_CONTROL_API_ZONE_NAME: Zone
    GM_CONTROL_DIFF_IGNORE_CREATE: "true"
  }
}

control_api: #Component & {
  name: "control-api"
  isStatefulset: false
  image: =~"^docker.greymatter.io/(release|development)/gm-control-api:" & !~"latest$"
  ports: api: 5555
  env: {
    GM_CONTROL_API_ADDRESS: "0.0.0.0:5555"
    GM_CONTROL_API_USE_TLS: "false"
    GM_CONTROL_API_ZONE_NAME: Zone
    GM_CONTROL_API_ZONE_KEY: Zone
    GM_CONTROL_API_DISABLE_VERSION_CHECK: "false"
    GM_CONTROL_API_PERSISTER_TYPE: "redis"
    GM_CONTROL_API_REDIS_MAX_RETRIES: "50"
    GM_CONTROL_API_REDIS_RETRY_DELAY: "5s"
    GM_CONTROL_API_REDIS_HOST: Redis.host
    GM_CONTROL_API_REDIS_PORT: Redis.port
    GM_CONTROL_API_REDIS_DB: Redis.db
  }
  envFrom: GM_CONTROL_API_REDIS_PASS: {
    secretKeyRef: {
      name: "gm-redis-password"
      key: "password"
    }
  }
}

catalog: #Component & {
  name: "catalog"
  isStatefulset: false
  image: =~"^docker.greymatter.io/(release|development)/gm-catalog:" & !~"latest$"
  ports: api: 8080
  env: {
    SEED_FILE_PATH: "/app/seed/seed.yaml"
    SEED_FILE_FORMAT: "yaml"
    CONFIG_SOURCE: "redis"
    REDIS_MAX_RETRIES: "50"
    REDIS_RETRY_DELAY: "5s"
    REDIS_HOST: Redis.host
    REDIS_PORT: Redis.port
    REDIS_DB: Redis.db
  }
  envFrom: REDIS_PASS: {
    secretKeyRef: {
      name: "gm-redis-password"
      key: "password"
    }
  }
  volumeMounts: "catalog-seed": {
    mountPath: "/app/seed"
  }
  volumes: "catalog-seed": {
    configMap: {
      name: "catalog-seed"
      defaultMode: 420
    }
  }
  configMaps: "catalog-seed": "seed.yaml": """
    \(MeshName):
      mesh_type: greymatter
      sessions:
        default:
          url: control.\(InstallNamespace).svc:50000
          zone: \(Zone)
    """
}

dashboard: #Component & {
  name: "dashboard"
  isStatefulset: false
  image: =~"^docker.greymatter.io/(release|development)/gm-dashboard:" & !~"latest$"
  ports: app: 1337
  env: {
    BASE_URL: =~"^/services/dashboard/" & =~"/$"
    FABRIC_SERVER: =~"/services/catalog/" & =~"/$"
    CONFIG_SERVER: =~"/services/control-api/" & =~"/v1.0$"
    PROMETHEUS_SERVER: "/services/prometheus/latest/api/v1/"
    REQUEST_TIMEOUT: "15000"
    USE_PROMETHEUS: "true"
    DISABLE_PROMETHEUS_ROUTES_UI: "false"
    ENABLE_INLINE_DOCS: "true"
  }
}

jwt_security: #Component & {
  name: "jwt-security"
  isStatefulset: false
  image: =~"^docker.greymatter.io/(release|development)/gm-jwt-security:" & !~"latest$"
  ports: api: 3000
  env: {
    HTTP_PORT: "3000"
    REDIS_HOST: Redis.host
    REDIS_PORT: Redis.port
    REDIS_DB: Redis.db
    ENABLE_TLS: "false" // TEMP!
  }
  envFrom: {
    REDIS_PASS: {
      secretKeyRef: {
        name: "gm-redis-password"
        key: "password"
      }
    }
    JWT_API_KEY: {
      secretKeyRef: {
        name: "jwt-keys"
        key: "api-key"
      }
    }
    PRIVATE_KEY: {
      secretKeyRef: {
        name: "jwt-keys"
        key: "private-key"
      }
    }
  }
  volumeMounts: "jwt-users": {
    mountPath: "/gm-jwt-security/etc"
  }
  volumes: "jwt-users": {
    configMap: {
      name: "jwt-users"
      defaultMode: 420
    }
  }
  configMaps: "jwt-users": "users.json": JWT.userTokens
  secrets: {
    "jwt-keys": {
      "api-key": JWT.apiKey
      "private-key": JWT.privateKey
    }
  }
}

redis: #Component & {
  name: "gm-redis"
  isStatefulset: true
  image: =~"redis:"
  command: ["redis-server"]
  args: [
    "--appendonly",
    "yes",
    "--requirepass",
    "$(REDIS_PASSWORD)",
    "--dir",
    "/data"
  ]
  ports: redis: 6379
  envFrom: REDIS_PASSWORD: {
    secretKeyRef: {
      name: "gm-redis-password"
      key: "password"
    }
  }
  secrets: "gm-redis-password": "password": Redis.password
  volumeMounts: "gm-redis-append-dir": {
    mountPath: "/data"
  }
  persistentVolumeClaims: "gm-redis-append-dir": {
    accessModes: ["ReadWriteOnce"]
    resources: requests: storage: "40Gi"
    volumeMode: "Filesystem"
  }
}

// TODO: Not currently being installed
prometheus: #Component & {
  name: "gm-prometheus"
  isStatefulset: true
  image: =~"^prom/prometheus:"
  command: ["/bin/prometheus"]
  args: [
    "--query.timeout=4m",
    "--query.max-samples=5000000000",
    "--storage.tsdb.path=/var/lib/prometheus/data/data",
    "--config.file=/etc/prometheus/prometheus.yaml",
    "--web.console.libraries=/usr/share/prometheus/console_libraries",
    "--web.console.templates=/usr/share/prometheus/consoles",
    "--web.enable-admin-api",
    "--web.external-url=http://anything/services/prometheus/latest", // TODO
    "--web.route-prefix=/"
  ]
  ports: prom: 9090
}
