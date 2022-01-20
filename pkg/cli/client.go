package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
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

	f, _ := fabric.New(mesh)

	cl := &client{
		mesh:        mesh.Name,
		flags:       flags,
		controlCmds: make(chan cmd),
		catalogCmds: make(chan cmd),
		ctx:         ctxt,
		cancel:      cancel,
		f:           f,
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
				}).run(cl.flags); err != nil {
					logger.Info("Waiting to connect to Control API", "Mesh", mesh.Name)
					time.Sleep(time.Second * 10)
					continue PING_CONTROL_LOOP
				}
				logger.Info("Connected to Control API",
					"Mesh", mesh.Name,
					"Elapsed", time.Since(start).String())
				break PING_CONTROL_LOOP
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
					logger.Info("requeued failed command", "args", c.args)
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
				}).run(cl.flags); err != nil {
					logger.Info("Waiting to connect to Catalog API", "Mesh", mesh.Name)
					time.Sleep(time.Second * 10)
					continue PING_CATALOG_LOOP
				}
				logger.Info("Connected to Catalog API",
					"Mesh", mesh.Name,
					"Elapsed", time.Since(start).String())
				break PING_CATALOG_LOOP
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
					logger.Info("requeued failed command", "args", c.args)
					catalogCmds <- c
				}
			}
		}
	}(cl.ctx, cl.catalogCmds)

	return cl
}
