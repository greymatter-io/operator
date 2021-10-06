package base

InstallNamespace: string

Zone: *"default-zone" | string

ImagePullSecretName: *"gm-docker-secret" | string

// TODO: Make option for applying ProxyPort
ProxyPort: *10808 | int32

Spire: *false | bool

UserTokens: *"[]" | string

Redis: {
  host: *"gm-redis.\(InstallNamespace).svc.cluster.local" | string
  port: *"6379" | string
  password: "" | string
  db: *"0" | string
}
