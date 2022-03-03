// All k8s manifests objects for core componenents drawn together
// for simultaneous application

package only

import "encoding/yaml"

// greymatter_namespace is currently in intermediates TODO move
k8s_manifests: controlensemble + edge + dashboard
// for CLI convenience
k8s_manifests_yaml: yaml.MarshalStream(k8s_manifests)

// TODO this was only necessary because I don't know how to pass _Name into #sidecar_container_block
// from Go. Then I decided to kill two birds with one stone and also put the sidecar_socket_volume in there.
// So for now, the way we get sidecar config for injected sidecars is to pull this structure and then
// separately apply the container and volumes to an intercepted Pod.
sidecar_container: {
  name: string
  
  container: #sidecar_container_block & {_Name: name}
  volumes: #spire_socket_volumes
}