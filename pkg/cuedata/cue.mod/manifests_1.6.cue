package onesix

proxy: {
  image: "docker.greymatter.io/release/gm-proxy:1.6.3"
}

control: {
  image: "docker.greymatter.io/release/gm-control:1.6.5"
}

control_api: {
  image: "docker.greymatter.io/release/gm-control-api:1.6.5"
}

catalog: {
  image: "docker.greymatter.io/release/gm-catalog:2.0.1"
}

dashboard: {
  image: "docker.greymatter.io/release/gm-dashboard:5.1.1"
  env: {
    BASE_URL: "/services/dashboard/"
    FABRIC_SERVER: "/services/catalog/"
    CONFIG_SERVER: "/services/control/api/v1.0"
    PROMETHEUS_SERVER: "/services/prometheus/api/v1/"
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
