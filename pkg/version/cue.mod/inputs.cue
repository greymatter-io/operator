package base

Namespace: string

// TODO: Make option for applying ProxyPort
ProxyPort: *10808 | int32

Spire: *false | bool

Redis: {
  external: *false | bool
  host: string
  port: string
  password: string
  db: string | "0"
}
