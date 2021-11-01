package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cuelang.org/go/cue"
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

func newClient(mesh *v1alpha1.Mesh, options []cue.Value, flags ...string) *client {
	ctxt, cancel := context.WithCancel(context.Background())

	cl := &client{
		mesh:        mesh.Name,
		flags:       flags,
		controlCmds: make(chan cmd),
		catalogCmds: make(chan cmd),
		ctx:         ctxt,
		cancel:      cancel,
		f:           fabric.New(options),
	}

	// Consume commands to send to Control
	go func(ctx context.Context, controlCmds chan cmd) {

		// Ping Control every 5s until responsive by getting and editing the Mesh's zone.
		// This ensures we can read and write from Control without any errors.
	PING_CONTROL_LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if r := (cmd{
					// args: fmt.Sprintf("get zone --zone-key %s", mesh.Spec.Zone),
					args: fmt.Sprintf("get zone %s", mesh.Spec.Zone),
					and: cmdOpt{cmd: &cmd{
						args:   fmt.Sprintf("edit zone %s", mesh.Spec.Zone),
						reader: values("zone_key"),
					}},
				}).run(cl.flags); r.err == nil {
					break PING_CONTROL_LOOP
				}
				time.Sleep(time.Second * 5)
			}
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
				// Requeue failed commands, since there are likely object dependencies (TODO: check)
				if r := c.run(cl.flags); r.err != nil && c.requeue {
					controlCmds <- c
				}
			}
		}
	}(cl.ctx, cl.controlCmds)

	// Consume commands to send to Catalog
	go func(ctx context.Context, catalogCmds chan cmd) {

		// Ping Catalog every 5s until responsive (getting the Mesh's session status with Control).
	PING_CATALOG_LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if r := (cmd{
					// args: fmt.Sprintf("get catalogmesh --mesh-id %s", mesh.Name),
					args:   fmt.Sprintf("get catalog-mesh %s", mesh.Name),
					reader: values("mesh_id", "session_statuses.default"),
				}).run(cl.flags); r.err == nil {
					break PING_CATALOG_LOOP
				}
				time.Sleep(time.Second * 5)
			}
		}

		logger.Info("Connected to Catalog", "Mesh", mesh.Name)

		// Then consume additional commands for catalog objects
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-catalogCmds:
				// Requeue failed commands, since there are likely object dependencies (TODO: check)
				if r := c.run(cl.flags); r.err != nil && c.requeue {
					catalogCmds <- c
				}
			}
		}
	}(cl.ctx, cl.catalogCmds)

	return cl
}

func mkApply(kind string, data json.RawMessage) cmd {
	kk := kindKey(kind)
	return cmd{
		args:    fmt.Sprintf("create %s", kind),
		requeue: true,
		stdin:   data,
		reader:  values(kk),
		or: cmdOpt{
			cmd: &cmd{
				args:    fmt.Sprintf("edit %s %s", kind, objKey(kind, data)),
				requeue: true,
				stdin:   data,
				reader:  values(kk),
			},
			when: func(out string) bool {
				return strings.Contains(out, "duplicate") || strings.Contains(out, "exists")
			},
		},
	}
}

func mkDelete(kind string, data json.RawMessage) cmd {
	key := objKey(kind, data)
	return cmd{args: fmt.Sprintf("delete %s %s", kind, key)}
}

func values(keys ...string) func(string) result {
	return func(out string) result {
		var kvs []interface{}
		for _, key := range keys {
			value := gjson.Get(out, key)
			if value.Exists() {
				// Add the gjson.Result without parsing its type.
				kvs = append(kvs, key, value)
			}
		}
		r := result{out, kvs, nil}
		if len(kvs) == 0 {
			r.err = fmt.Errorf("failed to get %v", keys)
		}
		return r
	}
}

func objKey(kind string, data json.RawMessage) string {
	result := values(kindKey(kind))(string(data))
	if len(result.kvs) != 2 {
		logger.Error(fmt.Errorf(kind), "no object key", "data", string(data))
		return ""
	}
	// The key value is a gjson.Result, so just format into a string.
	return fmt.Sprintf("%v", result.kvs[1])
}

func kindKey(kind string) string {
	if kind == "catalog-service" {
		return "service_id"
	}
	return fmt.Sprintf("%s_key", kind)
}
