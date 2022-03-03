// Contains nothing but the keys that need values from the outside. It is valuable to preserve the nearly flat, no-computation nature of this file

package only

import (
  corev1 "k8s.io/api/core/v1"
  "github.com/greymatter-io/operator/api/v1alpha1"
)

mesh: v1alpha1.#Mesh & {
  metadata: {
    name: string | *"mymesh"
  }
  spec: {
    install_namespace: string | *"greymatter"
    watch_namespaces: [...string] | *["default"]
    release_version: string | *"1.7"
    zone: string | *"default-zone"
    images: {
      proxy: string | *"docker.greymatter.io/release/gm-proxy:1.7.0-rc.4"
      catalog: string | *"docker.greymatter.io/release/gm-catalog:3.0.0-rc.3"
      dashboard: string | *"docker.greymatter.io/release/gm-dashboard:6.0.0-rc.2"

      control: string | *"docker.greymatter.io/internal/gm-control:1.7.1"
      control_api: string | *"docker.greymatter.io/internal/gm-control-api:1.7.1"

      redis: string | *"redis:latest"
    }
  }
}

defaults: {
  image_pull_secret_name: string | *"gm-docker-secret"
  image_pull_policy: corev1.#enumPullPolicy | *corev1.#PullAlways
  xds_host: "controlensemble.\(mesh.spec.install_namespace).svc.cluster.local"

  ports: {
    default_ingress: 10808
    redis_ingress: 10910
  }

}