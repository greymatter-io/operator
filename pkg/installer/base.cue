#cv: {
  image: string
  command?: string
  args?: [...string]
  resources?: #resources
  labels: [string]: string
  ports: [string]: int32
  envs: [string]: string
  envsFrom: [string]: #envSource
  volumes: [string]: #volumeSource
  volumeMounts: [string]: #volume
}

#resources: {
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
#envSource: {
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
#volumeSource: {
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
  // Name omitted; derived from key in map
  secret: {
    defaultMode?: int32
    optional?: bool
  }
} | {
  // Name omitted; derived from key in map
  persistentVolumeClaim: {
    readOnly?: bool
  }
} | {
  // Name omitted; derived from key in map
  configMap: {
    defaultMode?: int32
    optional?: bool
  }
}

#keyToPath: {
  key: string
  path: string
  mode?: int32
}

#volume: {
  mountPath: string
  readOnly?: bool
  subPath?: string
  mountPropagationMode?: string
  subPathExpr?: string
}

proxy: #cv & {
  image: =~"^docker.greymatter.io/(release|development)/gm-proxy:"
  ports: {
    "metrics": 8081
  }
  envs: {
    "ENVOY_ADMIN_LOG_PATH"?: _
    "PROXY_DYNAMIC": "true"
    "XDS_CLUSTER": ""
    "XDS_ZONE": ""
    "XDS_HOST": ""
    "XDS_PORT": "50000"
  }
}

edge: #cv & {
  image: ""
  envs: {
    "XDS_CLUSTER": "edge"
  }
}

control: #cv & {
  image: =~"^docker.greymatter.io/(release|development)/gm-control:"
  ports: {
    "xDS": 50000
  }
}

control_api: #cv & {}

catalog: #cv & {}

dashboard: #cv & {}

jwt_security: #cv & {}

redis: #cv & {}

prometheus: #cv & {}
