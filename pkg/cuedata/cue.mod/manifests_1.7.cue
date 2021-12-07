package oneseven

proxy: {
  image: "docker.greymatter.io/development/gm-proxy:1.7.0-rc.4"
}

control: {
  image: "docker.greymatter.io/development/gm-control:1.7.0-rc.3"
}

control_api: {
  image: "docker.greymatter.io/development/gm-control-api:1.7.0-rc.3"
}

catalog: {
  image: "docker.greymatter.io/development/gm-catalog:3.0.0-rc.3"
}

dashboard: {
  image: "docker.greymatter.io/development/gm-dashboard:6.0.0-rc.2"
  env: {
    CONFIG_SERVER: "/services/control/api/v1.0"
  }
}

jwt_security: {
  image: "docker.greymatter.io/development/gm-jwt-security:1.3.0"
}

redis: {
  image: "bitnami/redis:5.0.12"
}

prometheus: {
  image: "prom/prometheus:v2.7.1"
}
