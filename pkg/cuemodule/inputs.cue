// Contains keys that need values from the outside (i.e., from the Go) and defaults for each.
// Intended to be the single place where values are injected from the Go, though it overlaps a bit with config.cue.

package only

import (
  corev1 "k8s.io/api/core/v1"
  "github.com/greymatter-io/operator/api/v1alpha1"
)

config: {
  // Flags
  spire: bool | *true // enable Spire-based mTLS DEBUG - the default should be false
  auto_apply_mesh: bool | *true // apply the default mesh specified above after a delay
  generate_webhook_certs: bool | *true

  // Values
  cluster_ingress_name: "cluster" // For OpenShift deployments, this is used to look up the configured ingress domain
}

mesh: v1alpha1.#Mesh & {
  metadata: {
    name: string | *"mymesh"
  }
  spec: {
    install_namespace: string | *"greymatter"
    watch_namespaces: [...string] | *["default"]
    release_version: string | *"1.7" // no longer does anything, for the moment
    zone: string | *"default-zone"
    images: { // TODO start with defaults from below
      proxy: string | *"docker.greymatter.io/release/gm-proxy:1.7.0"
      catalog: string | *"docker.greymatter.io/release/gm-catalog:3.0.0"
      dashboard: string | *"docker.greymatter.io/release/gm-dashboard:6.0.0"

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
  redis_cluster_name: "redis"
  redis_host: "\(redis_cluster_name).\(mesh.spec.install_namespace).svc.cluster.local"

  // as new sidecars need to beacon metrics to Redis, this list will be updated dynamically
  // it is used in gm/outputs/redis.cue
  // TODO I don't like that this gets *read* by the Go, but it's not in EXTRACTME
  redis_spire_subjects: [...string] | *["dashboard", "catalog", "controlensemble", "edge"]


  ports: {
    default_ingress: 10808
    redis_ingress: 10910
  }

  images: {
    // TODO this is not the default image we actually want for 1.0, so update this later
    operator: string | *"docker.greymatter.io/internal/gm-operator:local_refactored" @tag(operator_image) // cibuild uses the tag
  }

}