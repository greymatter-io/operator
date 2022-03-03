package only

import (
	corev1 "k8s.io/api/core/v1"
)


// Note that this container block needs to go in `containers`, and refers to a
// spire-socket volume that must be configured separately (see below)
#sidecar_container_block: {
  _Name: string

  name: "sidecar"
  image: mesh.spec.images.proxy
  ports: [...corev1.#ContainerPort] | *[{
    name: "proxy"
    containerPort: defaults.ports.default_ingress
  }]
  env: [
    {name: "XDS_CLUSTER", value: _Name},
    {name: "ENVOY_ADMIN_LOG_PATH", value: "/dev/stdout"},
    {name: "PROXY_DYNAMIC", value: "true"},
    {name: "XDS_ZONE", value: mesh.spec.zone},
    {name: "XDS_HOST", value: defaults.xds_host},
    {name: "XDS_PORT", value: "50000"},
    {name: "SPIRE_PATH", value: "/run/spire/socket/agent.sock"}
  ]
  volumeMounts: [{
    name: "spire-socket",
    mountPath: "/run/spire/socket" }
  ]
  imagePullPolicy: defaults.image_pull_policy
}

// volume for spire referenced by proxy container block above
#spire_socket_volumes: [{
  name: "spire-socket",
  hostPath: {path: "/run/spire/socket", type: "DirectoryOrCreate"}
}]
