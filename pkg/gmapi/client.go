package gmapi

import (
	"context"
	"fmt"
	"github.com/greymatter-io/operator/pkg/cuemodule"
	"time"

	"github.com/google/uuid"
	"github.com/greymatter-io/operator/api/v1alpha1"
)

type Client struct {
	mesh        string
	flags       []string
	ControlCmds chan Cmd
	CatalogCmds chan Cmd
	Ctx         context.Context
	Cancel      context.CancelFunc
}

func newClient(operatorCUE *cuemodule.OperatorCUE, mesh *v1alpha1.Mesh, flags ...string) (*Client, error) {

	ctxt, cancel := context.WithCancel(context.Background())

	client := &Client{
		mesh:        mesh.Name,
		flags:       flags,
		ControlCmds: make(chan Cmd),
		CatalogCmds: make(chan Cmd),
		Ctx:         ctxt,
		Cancel:      cancel,
	}

	// Apply core Grey Matter components from CUE
	// This just dumps them on the channel, so it will block until the consumer is ready
	go ApplyCoreMeshConfigs(client, operatorCUE)

	// Consumer of commands to send to Control
	go func(ctx context.Context, controlCmds chan Cmd) {
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
				if _, err := (Cmd{
					// Create a NOOP shared_rules object to ensure that we can write to Control.
					// Using `greymatter create` is required because `greymatter apply` does not exit with an error code on failed actions.
					args: fmt.Sprintf("create sharedrules --zone-key %s --shared-rules-key %s --name %s", mesh.Spec.Zone, srKey, srKey),
				}).run(client.flags); err != nil {
					logger.Info("Waiting to connect to Control API", "Mesh", mesh.Name, "Issue", err)
					time.Sleep(time.Second * 10)
					continue PING_CONTROL_LOOP
				}
				logger.Info("Connected to Control API",
					"Mesh", mesh.Name,
					"Elapsed", time.Since(start).String())
				break PING_CONTROL_LOOP
			}
		}

		// Then consume additional commands for control objects
		for {
			select {
			case <-ctx.Done():
				return
			case c := <-controlCmds:
				// Requeue failed commands, since there are likely object dependencies (TODO: check)
				if _, err := c.run(client.flags); err != nil && c.requeue {
					logger.Info("command failed, will reattempt in 10 seconds", "args", c.args)
					go func(args string) {
						time.Sleep(10 * time.Second)
						select {
						case <-ctx.Done():
							return
						default:
							logger.Info("requeued failed command", "args", args)
							controlCmds <- c
						}
					}(c.args)
				}
			}
		}
	}(client.Ctx, client.ControlCmds)

	// Consumer of commands to send to Catalog
	go func(ctx context.Context, catalogCmds chan Cmd) {
		start := time.Now()

		// Ping Catalog every 5s until responsive (getting the Mesh's session status with Control).
	PING_CATALOG_LOOP:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if _, err := (Cmd{
					args: fmt.Sprintf("get catalogmesh --mesh-id %s", mesh.Name),
				}).run(client.flags); err != nil {
					logger.Info("Waiting to connect to Catalog API", "Mesh", mesh.Name, "Issue", err)
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
				if _, err := c.run(client.flags); err != nil && c.requeue {
					logger.Info("command failed, will reattempt in 10 seconds", "args", c.args)
					go func(args string) {
						time.Sleep(10 * time.Second)
						select {
						case <-ctx.Done():
							return
						default:
							logger.Info("requeued failed command", "args", args)
							catalogCmds <- c
						}
					}(c.args)
				}
			}
		}
	}(client.Ctx, client.CatalogCmds)

	return client, nil
}

func ApplyCoreMeshConfigs(client *Client, operatorCUE *cuemodule.OperatorCUE) {
	// by this point, GM has already been unified with THE mesh this operator manages
	// Extract correct GM config for options - for now there's only one

	meshConfigs, kinds, err := operatorCUE.ExtractCoreMeshConfigs()
	if err != nil {
		logger.Error(err, "failed to extract while attempting to apply core components mesh config - ignoring")
		return
	}
	ApplyAll(client, meshConfigs, kinds)
}
