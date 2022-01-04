package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cuelang.org/go/cue"
	"github.com/google/uuid"
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
		start := time.Now()

		// Generate a random shared_rules object key to create a dummy object that ensures we can write to Control.
		srKey := uuid.New().String()

		// Ping Control every 5s until responsive by getting and editing the Mesh's zone.
		// This ensures we can read and write from Control without any errors.
	PING_CONTROL_LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if _, err := (cmd{
					// Create a NOOP shared_rules object to ensure that we can write to Control.
					// Using `greymatter create` is required because `greymatter apply` does not exit with an error code on failed actions.
					args: fmt.Sprintf("create sharedrules --zone-key %s --shared-rules-key %s --name %s", mesh.Spec.Zone, srKey, srKey),
					log: func(err error) {
						if err != nil {
							logger.Info("Waiting to connect to Control API", "Mesh", mesh.Name)
						} else {
							logger.Info("Connected to Control API",
								"Mesh", mesh.Name,
								"Elapsed", time.Since(start).String())
						}
					},
				}).run(cl.flags); err == nil {
					break PING_CONTROL_LOOP
				}
				time.Sleep(time.Second * 10)
			}
		}

		// Configure edge domain, since it is a dependency for all sidecar routes.
		mkApply(mesh.Name, "domain", cl.f.EdgeDomain()).run(cl.flags)

		// Then consume additional commands for control objects
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-controlCmds:
				// Requeue failed commands, since there are likely object dependencies (TODO: check)
				if _, err := c.run(cl.flags); err != nil && c.requeue {
					controlCmds <- c
				}
			}
		}
	}(cl.ctx, cl.controlCmds)

	// Consume commands to send to Catalog
	go func(ctx context.Context, catalogCmds chan cmd) {
		start := time.Now()

		// Ping Catalog every 5s until responsive (getting the Mesh's session status with Control).
	PING_CATALOG_LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if _, err := (cmd{
					args: fmt.Sprintf("get catalogmesh --mesh-id %s", mesh.Name),
					log: func(err error) {
						if err != nil {
							logger.Info("Waiting to connect to Catalog API", "Mesh", mesh.Name)
						} else {
							logger.Info("Connected to Catalog API",
								"Mesh", mesh.Name,
								"Elapsed", time.Since(start).String())
						}
					},
				}).run(cl.flags); err == nil {
					break PING_CATALOG_LOOP
				}
				time.Sleep(time.Second * 10)
			}
		}

		// Then consume additional commands for catalog objects
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-catalogCmds:
				// Requeue failed commands, since there are likely object dependencies (TODO: check)
				if _, err := c.run(cl.flags); err != nil && c.requeue {
					catalogCmds <- c
				}
			}
		}
	}(cl.ctx, cl.catalogCmds)

	return cl
}

func mkApply(mesh, kind string, data json.RawMessage) cmd {
	key := objKey(kind, data)
	return cmd{
		args:    fmt.Sprintf("apply -t %s -f -", kind),
		requeue: true,
		stdin:   data,
		log: func(err error) {
			if err != nil {
				logger.Error(err, "failed apply")
			} else {
				logger.Info("apply", "type", kind, "key", key, "Mesh", mesh)
			}
		},
	}
}

func mkDelete(mesh, kind string, data json.RawMessage) cmd {
	key := objKey(kind, data)
	args := fmt.Sprintf("delete %s --%s %s", kind, kindFlag(kind), key)
	if kind == "catalogservice" {
		args += fmt.Sprintf(" --mesh-id %s", mesh)
	}
	return cmd{
		args: args,
		log: func(err error) {
			if err != nil {
				logger.Error(err, "failed delete")
			} else {
				logger.Info("delete", "type", kind, "key", key, "Mesh", mesh)
			}
		},
	}
}

func objKey(kind string, data json.RawMessage) string {
	key := kindKey(kind)
	value := gjson.Get(string(data), key)
	if value.Exists() {
		return value.String()
	}
	logger.Error(fmt.Errorf(kind), "no object key", "data", string(data))
	return ""
}

func kindKey(kind string) string {
	if kind == "catalogservice" {
		return "service_id"
	}
	return fmt.Sprintf("%s_key", kind)
}

func kindFlag(kind string) string {
	if kind == "catalogservice" {
		return "service-id"
	}
	return fmt.Sprintf("%s-key", kind)
}
