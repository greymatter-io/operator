package base

// All known versions are lists of known OCI image strings with an optional
// override from an applied mesh CRD.
Versions: {
  if mesh.spec.release_version == "latest" {
    proxy: *"docker.greymatter.io/raw/development/gm-proxy:latest" | mesh.spec.images["proxy"]
    control: *"docker.greymatter.io/raw/development/gm-control:latest" | mesh.spec.images["control"]
    control_api: *"docker.greymatter.io/raw/development/gm-control-api:latest" | mesh.spec.images["control_api"]
    catalog: *"docker.greymatter.io/raw/development/gm-catalog:latest" | mesh.spec.images["catalog"]
    dashboard: *"docker.greymatter.io/raw/development/gm-dashboard:latest" | mesh.spec.images["dashboard"]
    jwtsecurity: *"docker.greymatter.io/raw/development/gm-jwt-security:latest" | mesh.spec.images["jwtsecurity"]
    redis: *"bitnami/redis:5.0.12" | mesh.spec.images["redis"]
    prometheus: *"prom/prometheus:v2.7.1" | mesh.spec.images["prometheus"]
  }
  
  if mesh.spec.release_version == "1.7" {
    proxy: *"docker.greymatter.io/release/gm-proxy:1.6.3" | mesh.spec.images["proxy"]
    control: *"docker.greymatter.io/release/gm-proxy:1.7.0" | mesh.spec.images["control"]
    control_api: *"docker.greymatter.io/release/gm-proxy:1.7.0" | mesh.spec.images["control_api"]
    catalog: *"docker.greymatter.io/release/gm-catalog:3.0.0" | mesh.spec.images["catalog"]
    dashboard: *"docker.greymatter.io/release/gm-dashboard:6.0.0" | mesh.spec.images["dashboard"]
    jwtsecurity: *"dockery.greymatter.io/release/gm-jwt-security:1.3.0" | mesh.spec.images["jwtsecurity"]
    redis: *"bitnami/redis:5.0.12" | mesh.spec.images["redis"]
    prometheus: *"prom/prometheus:v2.7.1" | mesh.spec.images["prometheus"]
  }

  if mesh.spec.release_version == "1.6" {
    proxy: *"docker.greymatter.io/release/gm-proxy:1.6.3" | mesh.spec.images["proxy"]
    control: *"docker.greymatter.io/release/gm-control:1.6.5" | mesh.spec.images["control"]
    control_api: *"docker.greymatter.io/release/gm-control-api:1.6.5" | mesh.spec.images["control_api"]
    catalog: *"docker.greymatter.io/release/gm-catalog:2.0.1" | mesh.spec.images["catalog"]
    dashboard: *"docker.greymatter.io/release/gm-dashboard:5.1.1" | mesh.spec.images["dashboard"]
    jwtsecurity: *"dockery.greymatter.io/release/gm-jwt-security:1.3.0" | mesh.spec.images["jwtsecurity"]
    redis: *"bitnami/redis:5.0.12" | mesh.spec.images["redis"]
    prometheus: *"prom/prometheus:v2.7.1" | mesh.spec.images["prometheus"]
  }
}
