package base

MeshName: string

// Where to install components
InstallNamespace: string
// The scope of the mesh network; includes InstallNamespace
WatchNamespaces: string

MeshName: string

ClusterType: *"openshift" | string
IngressSubDomain: *"localhost" | string

Zone: *"default-zone" | string

ImagePullSecretName: *"gm-docker-secret" | string

MeshPort: *10808 | int32

Spire: *false | bool

JWT: {
  userTokens: *"[]" | string
  apiKey: "" | string
  privateKey: "" | string
}

Redis: {
  host: *"gm-redis.\(InstallNamespace).svc.cluster.local" | string
  port: *"6379" | string
  password: "" | string
  db: *"0" | string
}
