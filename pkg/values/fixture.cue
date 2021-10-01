
proxy: {
  image: "docker.greymatter.io/release/gm-proxy:1.6.4"
  ports: {
    metrics: 8081
  }
  envs: {
    ENVOY_ADMIN_LOG_PATH: "/dev/stdout"
    PROXY_DYNAMIC: "true"
    XDS_PORT: "50000"
  }
}

edge: {
  envs: {
    XDS_CLUSTER: "edge"
  }
}

control: {
  image: "docker.greymatter.io/release/gm-control:1.6.4"
  ports: {
    grpc: 50000
  }
  envs: {
    GM_CONTROL_CMD: "kubernetes"
    GM_CONTROL_KUBERNETES_CLUSTER_LABEL: "greymatter.io/cluster"
    GM_CONTROL_KUBERNETES_PORT_NAME: "proxy"
    GM_CONTROL_API_HOST: "127.0.0.1:5555" // share one deployment!
    GM_CONTROL_API_INSECURE: "true"
    GM_CONTROL_API_SSL: "false"
  }
}

control_api: {
  image: "docker.greymatter.io/release/gm-control-api:1.6.4"
  labels: {
    "greymatter.io/cluster": "control-api"
  }
  ports: {
    api: 5555
  }
  envs: {
    GM_CONTROL_API_ADDRESS: "0.0.0.0:5555"
    GM_CONTROL_API_DISABLE_VERSION_CHECK: "false"
    GM_CONTROL_API_PERSISTER_TYPE: "redis"
    GM_CONTROL_API_REDIS_MAX_RETRIES: "50"
    GM_CONTROL_API_REDIS_RETRY_DELAY: "5s"
  }
}

catalog: {
  image: "docker.greymatter.io/release/gm-catalog:2.0.0"
  labels: {
    "greymatter.io/cluster": "catalog"
  }
  ports: {
    api: 8080
  }
  envs: {
    CONFIG_SOURCE: "redis"
    REDIS_MAX_RETRIES: "50"
    REDIS_RETRY_DELAY: "5s"
  }
}

dashboard: {
  image: "docker.greymatter.io/release/gm-dashboard:5.0.0"
  labels:
    "greymatter.io/cluster": "dashboard"
  ports: {
    app: 1337
  }
  envs: {
    REQUEST_TIMEOUT: "15000"
    BASE_URL: "/services/dashboard/5.0/"
    FABRIC_SERVER: "/services/catalog/2.0/"
    CONFIG_SERVER: "/services/control-api/1.6/v1.0"
    PROMETHEUS_SERVER: "/services/prometheus/latest/api/v1/"
    USE_PROMETHEUS: "true"
    DISABLE_PROMETHEUS_ROUTES_UI: "false"
    ENABLE_INLINE_DOCS: "true"
  }
}

jwt_security: {
  image: "docker.greymatter.io/release/gm-jwt-security:1.3.0"
  labels:
    "greymatter.io/cluster": "jwt-security"
  ports: {
    api: 3000
  }
  envs: {
    HTTP_PORT: "3000"
  }
  volumes: {
    "jwt-users": {
      configMap: {
        defaultMode: 420
      }
    }
  }
  volumeMounts: {
    "jwt-users": {
      mountPath: "/gm-jwt-security/etc"
    }
  }
}

redis: {
  image: "bitnami/redis:5.0.12"
  command: "redis-server"
  args: [
    "--appendonly",
    "yes",
    "--requirepass",
    "$(REDIS_PASSWORD)",
    "--dir",
    "/data"
  ]
  ports: {
    redis: 6379
  }
}

prometheus: {
  image: "prom/prometheus:v2.7.1"
}
