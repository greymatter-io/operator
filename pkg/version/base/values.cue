package base

#Values: {
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
