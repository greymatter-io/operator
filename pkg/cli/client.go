package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	"github.com/tidwall/gjson"
)

type client struct {
	mesh        string
	flags       []string
	controlCmds chan cmd
	catalogCmds chan cmd
	ctx         context.Context
	cancel      context.CancelFunc
	f           *fabric.Fabric
}

func newClient(mesh *v1alpha1.Mesh, flags ...string) *client {
	ctxt, cancel := context.WithCancel(context.Background())

	cl := &client{
		mesh:        mesh.Name,
		flags:       flags,
		controlCmds: make(chan cmd),
		catalogCmds: make(chan cmd),
		ctx:         ctxt,
		cancel:      cancel,
		f:           fabric.New(mesh),
	}

	// Consume commands to send to Control
	go func(ctx context.Context, controlCmds chan cmd) {

		// Ping Control every 5s until responsive by getting and editing the Mesh's zone.
		// This ensures we can read and write from Control without any errors.
		if (cmd{
			// args: fmt.Sprintf("get zone --zone-key %s", mesh.Spec.Zone),
			args: fmt.Sprintf("get zone %s", mesh.Spec.Zone),
			retry: retry{
				dur:  time.Second * 5,
				done: ctx.Done,
			},
			and: cmdOpt{cmd: &cmd{
				args:   fmt.Sprintf("edit zone %s", mesh.Spec.Zone),
				reader: values("zone_key", "checksum"),
				retry: retry{
					dur:  time.Second * 5,
					done: ctx.Done,
				},
			}},
		}).run(cl.flags).err != nil {
			return
		}

		logger.Info("Connected to Control", "Mesh", mesh.Name)

		// Configure edge domain, since it is a dependency for all sidecar routes.
		mkApply("domain", cl.f.EdgeDomain()).run(cl.flags)

		// Then consume additional commands for control objects
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-controlCmds:
				c.run(cl.flags)
			}
		}
	}(cl.ctx, cl.controlCmds)

	// Consume commands to send to Catalog
	go func(ctx context.Context, catalogCmds chan cmd) {

		// Ping Catalog every 5s until responsive (getting the Mesh's session status with Control).
		if (cmd{
			// args: fmt.Sprintf("get catalogmesh --mesh-id %s", mesh.Name),
			args:   fmt.Sprintf("get catalog-mesh %s", mesh.Name),
			reader: values("mesh_id", "session_statuses.default"),
			retry: retry{
				dur:  time.Second * 5,
				done: ctx.Done,
			},
		}).run(cl.flags).err != nil {
			return
		}

		logger.Info("Connected to Catalog", "Mesh", mesh.Name)

		// Then consume additional commands for catalog objects
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-catalogCmds:
				c.run(cl.flags)
			}
		}
	}(cl.ctx, cl.catalogCmds)

	return cl
}

// temp while CLI 4 is being worked on
func mkApply(kind string, data json.RawMessage) cmd {
	var kindKey string
	if kind == "catalog-service" {
		// if kind == "catalogservice" {
		kindKey = "service_id"
	} else {
		kindKey = fmt.Sprintf("%s_key", kind)
	}
	key := values(kindKey)(string(data)).kvs[1]
	return cmd{
		args:   fmt.Sprintf("create %s", kind),
		stdin:  data,
		reader: values(kindKey, "checksum"),
		or: cmdOpt{
			cmd: &cmd{
				args:   fmt.Sprintf("edit %s %s", kind, key),
				stdin:  data,
				reader: values(kindKey, "checksum"),
			},
			when: func(out string) bool {
				return strings.Contains(out, "duplicate") || strings.Contains(out, "exists")
			},
		},
	}
}

func mkDelete(kind, key string) cmd {
	kindKey := fmt.Sprintf("%s_key", kind)
	return cmd{
		args:   fmt.Sprintf("delete %s %s", kind, key),
		reader: values(kindKey, "checksum"),
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
