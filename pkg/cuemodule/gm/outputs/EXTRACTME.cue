// All Grey Matter config objects for core componenents drawn together
// for simultaneous application

package only

import "encoding/yaml"

mesh_configs: redis_config + edge_config + catalog_config + controlensemble_config + dashboard_config + catalog_entries
redis_listener: redis_listener_object // special because we need to re-apply it when Spire is enabled for every new sidecar

// for CLI convenience,
// e.g. `cue eval -c ./gm/outputs --out text -e mesh_configs_yaml`
mesh_configs_yaml: yaml.MarshalStream(mesh_configs)

sidecar_config: #sidecar_config // pass a Name and Port
