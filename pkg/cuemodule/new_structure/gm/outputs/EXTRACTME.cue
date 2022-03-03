// All Grey Matter config objects for core componenents drawn together
// for simultaneous application

package only

import "encoding/yaml"

mesh_configs: edge_config + controlensemble_config + dashboard_config + catalog_entries
// for CLI convenience
mesh_configs_yaml: yaml.MarshalStream(mesh_configs)

sidecar_config: #sidecar_config // pass a Name and Port