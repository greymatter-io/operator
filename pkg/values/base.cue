#ContainerValues: {
  image?: string
  command?: string
  args?: [...string]
  resources?: #Resources
  labels?: [string]: string
  ports?: [string]: int32
  envs?: [string]: string 
  envsFrom?: [string]: #EnvSource
  volumes?: [string]: #VolumeSource
  volumeMounts?: [string]: #VolumeMount
}

#Resources: {
  limits?: {
    cpu?: string
    memory?: string
  }
  requests?: {
    cpu?: string
    memory?: string
  }
}

// https://pkg.go.dev/k8s.io/api/core/v1#EnvVarSource
#EnvSource: {
  fieldRef?: {
    apiVersion?: string
    fieldPath: string
  }
  resourceFieldRef?: {
    containerName?: string
    resource: string
    divisor: string
  }
  configMapKeyRef?: {
    name: string
    key: string
    optional?: bool
  }
  secretKeyRef?: {
    name: string
    key: string
    optional?: bool
  }
}

// TODO: https://pkg.go.dev/k8s.io/api/core/v1#VolumeSource
#VolumeSource: {
  hostPath: {  
    path: string
    type?: string
  }
} | {
  emptyDir: {
    medium?: string
    sizeLimit?: string
  }
} | {
  secret: {
    secretName?: string // optional since we set it at runtime
    defaultMode?: int32
    optional?: bool
  }
} | {
  persistentVolumeClaim: {
    claimName?: string // optional since we set it at runtime
    readOnly?: bool
  }
} | {
  configMap: {
    name?: string // optional since we set it at runtime
    defaultMode?: int32
    optional?: bool
  }
}

#VolumeMount: {
  name?: string // optional since we set it at runtime
  mountPath: string
  readOnly?: bool
  subPath?: string
  mountPropagationMode?: string
  subPathExpr?: string
}

proxy: #ContainerValues & {
  image: =~"^docker.greymatter.io/(release|development)/gm-proxy:" & !~"latest$"
  ports: {
    metrics: 8081
  }
  envs: {
    ENVOY_ADMIN_LOG_PATH: "/dev/stdout"
    PROXY_DYNAMIC: "true"
    XDS_PORT: "50000"
  }
}

edge: #ContainerValues & {
  envs: {
    XDS_CLUSTER: "edge"
  }
}

control: #ContainerValues & {
  image: =~"^docker.greymatter.io/(release|development)/gm-control:" & !~"latest$"
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

control_api: #ContainerValues & {
  image: =~"^docker.greymatter.io/(release|development)/gm-control-api:" & !~"latest$"
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

catalog: #ContainerValues & {
  image: =~"^docker.greymatter.io/(release|development)/gm-catalog:" & !~"latest$"
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

dashboard: #ContainerValues & {
  image: =~"^docker.greymatter.io/(release|development)/gm-dashboard:" & !~"latest$"
  labels:
    "greymatter.io/cluster": "dashboard"
  ports: {
    app: 1337
  }
  envs: {
    REQUEST_TIMEOUT: "15000"
    BASE_URL: =~"^/services/dashboard/" & =~"/$"
    FABRIC_SERVER: =~"/services/catalog/" & =~"/$"
    CONFIG_SERVER: =~"/services/control-api/" & =~"v1.0$"
    PROMETHEUS_SERVER: "/services/prometheus/latest/api/v1/"
    USE_PROMETHEUS: "true"
    DISABLE_PROMETHEUS_ROUTES_UI: "false"
    ENABLE_INLINE_DOCS: "true"
  }
}

jwt_security: #ContainerValues & {
  image: =~"^docker.greymatter.io/(release|development)/gm-jwt-security:" & !~"latest$"
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

redis: #ContainerValues & {
  image: =~"redis:"
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

prometheus: #ContainerValues & {
  image: =~"^prom/prometheus:"
}
