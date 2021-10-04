package base

import (
	corev1 "k8s.io/api/core/v1"
)

Namespace: string
proxyPort: int32

#Component: {
  name: string
  image: string
  command: [...string]
  args: [...string]
  resources: corev1.#Resources
  ports: [string]: int32
  envFrom: [string]: corev1.#EnvVarSource
  env: [string]: string
  volumeMounts: [string]: corev1.#VolumeMount
  volumes: [string]: corev1.#VolumeSource
}

proxy: #Component & {
  image: =~"^docker.greymatter.io/(release|development)/gm-proxy:" & !~"latest$"
  env: {
    ENVOY_ADMIN_LOG_PATH: "/dev/stdout",
    PROXY_DYNAMIC: "true"
    XDS_PORT: "50000"
  }
}

edge: #Component & proxy & {
  name: "edge"
  env: XDS_CLUSTER: "edge"
}

control: #Component & {
  name: "control"
  image: =~"^docker.greymatter.io/(release|development)/gm-control:" & !~"latest$"
  ports: grpc: 50000
  env: {
    GM_CONTROL_CMD: "kubernetes"
    GM_CONTROL_KUBERNETES_CLUSTER_LABEL: "greymatter.io/cluster"
    GM_CONTROL_KUBERNETES_PORT_NAME: "proxy"
    GM_CONTROL_API_HOST: "127.0.0.1:5555" // share one deployment!
    GM_CONTROL_API_INSECURE: "true"
    GM_CONTROL_API_SSL: "false"
  }
}

control_api: #Component & {
  name: "control-api"
  image: =~"^docker.greymatter.io/(release|development)/gm-control-api:" & !~"latest$"
  ports: api: 5555
  env: {
    GM_CONTROL_API_ADDRESS: "0.0.0.0:5555"
    GM_CONTROL_API_DISABLE_VERSION_CHECK: "false"
    GM_CONTROL_API_PERSISTER_TYPE: "redis"
    GM_CONTROL_API_REDIS_MAX_RETRIES: "50"
    GM_CONTROL_API_REDIS_RETRY_DELAY: "5s"
  }
}

catalog: #Component & {
  name: "catalog"
  image: =~"^docker.greymatter.io/(release|development)/gm-catalog:" & !~"latest$"
  ports: api: 8080
  env: {
    CONFIG_SOURCE: "redis"
    REDIS_MAX_RETRIES: "50"
    REDIS_RETRY_DELAY: "5s"
  }
}

dashboard: #Component & {
  name: "dashboard"
  image: =~"^docker.greymatter.io/(release|development)/gm-dashboard:" & !~"latest$"
  ports: app: 1337
  env: {
    BASE_URL: =~"^/services/dashboard/" & =~"/$"
    FABRIC_SERVER: =~"/services/catalog/" & =~"/$"
    CONFIG_SERVER: =~"/services/control-api/" & =~"v1.0$"
    PROMETHEUS_SERVER: "/services/prometheus/latest/api/v1/"
    REQUEST_TIMEOUT: "15000"
    USE_PROMETHEUS: "true"
    DISABLE_PROMETHEUS_ROUTES_UI: "false"
    ENABLE_INLINE_DOCS: "true"
  }
}

jwt_security: #Component & {
  name: "jwt-security"
  image: =~"^docker.greymatter.io/(release|development)/gm-jwt-security:" & !~"latest$"
  ports: api: 3000
  env: {
    HTTP_PORT: "3000"
  }
  volumeMounts: {
    "jwt-users": {
      mountPath: "/gm-jwt-security/etc"
    }
  }
  volumes: {
    "jwt-users": {
      configMap: {
        defaultMode: 420
      }
    }
  }
}

redis: #Component & {
  name: "greymatter-redis"
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
}

prometheus: #Component & {
  name: "greymatter-prometheus"
  image: =~"^prom/prometheus:"
}
