proxy: {
  image: "docker.greymatter.io/release/gm-proxy:1.7.0"
}

control: {
  image: "docker.greymatter.io/release/gm-control:1.7.0"
}

control_api: {
  image: "docker.greymatter.io/release/gm-control-api:1.7.0"
}

catalog: {
  image: "docker.greymatter.io/release/gm-catalog:3.0.0"
}

dashboard: {
  image: "docker.greymatter.io/release/gm-dashboard:6.0.0"
  env: {
    CONFIG_SERVER: "/services/control/api/v1.0"
  }
}

jwt_security: {
  image: "docker.greymatter.io/release/gm-jwt-security:1.3.0"
}

redis: {
  image: "bitnami/redis:5.0.12"
}

prometheus: {
  image: "prom/prometheus:v2.7.1"
}
