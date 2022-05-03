// All k8s manifests objects for core componenents drawn together
// for simultaneous application

package only

import "encoding/yaml"

spire_manifests: spire_namespace + spire_server + spire_agent
operator_manifests: operator_namespace + operator_crd + operator_sts + operator_k8s + [for x in spire_manifests if config.spire {x}] // HACK but cleaner than a pile of ifs
k8s_manifests: controlensemble + catalog + redis + edge + dashboard

// for CLI convenience,
// e.g. `cue eval -c ./k8s/outputs --out text -e k8s_manifests_yaml`
operator_manifests_yaml: yaml.MarshalStream(operator_manifests)
spire_manifests_yaml: yaml.MarshalStream(spire_manifests)
k8s_manifests_yaml: yaml.MarshalStream(k8s_manifests)

// TODO this was only necessary because I don't know how to pass _Name into #sidecar_container_block
// from Go. Then I decided to kill two birds with one stone and also put the sidecar_socket_volume in there.
// So for now, the way we get sidecar config for injected sidecars is to pull this structure and then
// separately apply the container and volumes to an intercepted Pod.
sidecar_container: {
  name: string | *"REPLACEME" // has a default just so literally everything is concrete by default
  
  container: #sidecar_container_block & {_Name: name}
  volumes: #spire_socket_volumes
}
