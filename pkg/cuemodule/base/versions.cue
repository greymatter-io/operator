package base

Versions: {
  for k, v in _versions {
    if mesh.spec.images[k] == _|_ {
      "\(k)": v
    }
    if mesh.spec.images[k] != _|_ {
      "\(k)": mesh.spec.images[k]
    }
  }
}

_versions: {
  redis: "bitnami/redis:5.0.12"
  prometheus: "prom/prometheus:v2.7.1"

  if mesh.spec.release_version == "latest" {
    proxy: "docker.greymatter.io/development/gm-proxy:latest"
    control: "docker.greymatter.io/development/gm-control:latest"
    control_api: "docker.greymatter.io/development/gm-control-api:latest"
    catalog: "docker.greymatter.io/development/gm-catalog:latest"
    dashboard: "docker.greymatter.io/development/gm-dashboard:latest"
    jwtsecurity: "docker.greymatter.io/development/gm-jwt-security:latest"
  }

  if mesh.spec.release_version == "1.7" {
    proxy: "docker.greymatter.io/release/gm-proxy:1.7.0-rc.4"
    control: "docker.greymatter.io/release/gm-control:1.7.0-rc.3"
    control_api: "docker.greymatter.io/release/gm-control-api:1.7.0-rc.3"
    catalog: "docker.greymatter.io/release/gm-catalog:3.0.0-rc.3"
    dashboard: "docker.greymatter.io/release/gm-dashboard:6.0.0-rc.2"
    jwtsecurity: "docker.greymatter.io/release/gm-jwt-security:1.3.0"
  }

  if mesh.spec.release_version == "1.6" {
    proxy: "docker.greymatter.io/release/gm-proxy:1.6.3"
    control: "docker.greymatter.io/release/gm-control:1.6.5"
    control_api: "docker.greymatter.io/release/gm-control-api:1.6.5"
    catalog: "docker.greymatter.io/release/gm-catalog:2.0.1"
    dashboard: "docker.greymatter.io/release/gm-dashboard:5.1.1"
    jwtsecurity: "docker.greymatter.io/release/gm-jwt-security:1.3.0"
  }
}
