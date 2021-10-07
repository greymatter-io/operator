package base

Namespace: string

MeshPort: *10808 | int32

Spire: *false | bool

Redis: {
  host: *"greymatter-redis.\(Namespace).svc.cluster.local" | string
  port: *"6379" | string
  password: "" | string
  db: *"0" | string
}

WatchNamespaces: string