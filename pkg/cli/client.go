package cli

import (
	"encoding/json"
	"fmt"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	"github.com/tidwall/gjson"
)

type client struct {
	mesh        string
	flags       []string
	controlCmds chan cmd
	catalogCmds chan cmd
	f           *fabric.Fabric
}

func newClient(mesh *v1alpha1.Mesh, flags ...string) *client {
	cl := &client{
		mesh:        mesh.Name,
		flags:       flags,
		controlCmds: make(chan cmd),
		catalogCmds: make(chan cmd),
		f:           fabric.New(mesh),
	}

	// Consume commands to send to Control
	go func(controlCmds chan cmd) {
		// Ping Control every 5s until responsive (getting the Mesh's zone's checksum)
		(cmd{
			// args: fmt.Sprintf("get zone --zone-key %s", mesh.Spec.Zone),
			args:   fmt.Sprintf("get zone %s", mesh.Spec.Zone),
			reader: values("zone_key", "checksum"),
		}).persist(5, cl.flags)
		logger.Info("Connected to Control", "Mesh", mesh.Name)

		// Configure edge objects
		objects := cl.f.Edge()

		for _, c := range []cmd{
			mkApply("domain", objects.Domain),
			mkApply("listener", objects.Listener),
			mkApply("proxy", objects.Proxy),
			mkApply("cluster", objects.Cluster),
		} {
			c.run(cl.flags)
		}

		// Then consume additional commands for control objects
		for c := range controlCmds {
			c.run(cl.flags)
		}
	}(cl.controlCmds)

	// Consume commands to send to Catalog
	go func(catalogCmds chan cmd) {
		// Ping Catalog every 5s until responsive (getting the Mesh's session status with Control).
		(cmd{
			// args: fmt.Sprintf("get catalogmesh --mesh-id %s", mesh.Name),
			args:   fmt.Sprintf("get catalog-mesh %s", mesh.Name),
			reader: values("mesh_id", "session_statuses.default"),
		}).persist(5, cl.flags)
		logger.Info("Connected to Catalog", "Mesh", mesh.Name)

		// Then consume additional commands for catalog objects
		for c := range catalogCmds {
			c.run(cl.flags)
		}
	}(cl.catalogCmds)

	return cl
}

// temp while CLI 4 is being worked on
func mkApply(kind string, data json.RawMessage) cmd {
	kindKey := fmt.Sprintf("%s_key", kind)
	key := values(kindKey)(string(data)).kvs[1]
	return cmd{
		args:   fmt.Sprintf("create %s", kind),
		stdin:  data,
		reader: values(kindKey, "checksum"),
		backup: &cmd{
			args:   fmt.Sprintf("edit %s %s", kind, key),
			stdin:  data,
			reader: values(kindKey, "checksum"),
		},
	}
}

func values(keys ...string) func(string) result {
	return func(out string) result {
		var kvs []interface{}
		for _, key := range keys {
			value := gjson.Get(out, key)
			if value.Exists() {
				kvs = append(kvs, key, value)
			}
		}
		r := result{kvs, nil}
		if len(kvs) == 0 {
			r.err = fmt.Errorf("failed to get %v", keys)
		}
		return r
	}
}
