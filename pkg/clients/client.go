package clients

import (
	"fmt"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
)

type client struct {
	mesh    string
	flags   []string
	cmds    chan cmd
	results chan result
	f       *fabric.Fabric
}

func newClient(mesh *v1alpha1.Mesh, flags ...string) (*client, error) {
	f, err := fabric.New(mesh)
	if err != nil {
		return nil, err
	}

	cl := &client{
		mesh:    mesh.Name,
		flags:   flags,
		cmds:    make(chan cmd),
		results: make(chan result),
		f:       f,
	}

	// Start consuming the client's cmds channel.
	// The channel will close upon cleanup.
	go func(c *client) {
		for cmd := range c.cmds {
			out, err := cmd.run(c.flags)
			if err != nil {
				logger.Error(err, cmd.args, "Mesh", c.mesh)
			}
			c.results <- result{out: out, err: err}
		}
	}(cl)

	// Ping Control
	if !cl.retry(cmd{args: fmt.Sprintf("get zone %s", mesh.Spec.Zone)}, 5) {
		return nil, fmt.Errorf("failed to connect to Control")
	}
	logger.Info("Connected to Control", "Mesh", cl.mesh)

	// Ping Catalog
	if !cl.retry(cmd{args: fmt.Sprintf("get catalog-mesh %s", mesh.Name)}, 5) {
		return nil, fmt.Errorf("failed to connect to Catalog")
	}
	logger.Info("Connected to Catalog", "Mesh", cl.mesh)

	return cl, nil
}

func (cl *client) retry(c cmd, count int) bool {
	attempt := 1
	cl.cmds <- c
	for result := range cl.results {
		if result.err != nil {
			logger.Error(result.err, fmt.Sprintf("%s: retrying in 3s", c.args), "result", result.out, "Mesh", cl.mesh)
			time.Sleep(time.Second * 3)
			cl.cmds <- c
			attempt++
			if attempt == count {
				return false
			}
		} else {
			break
		}
	}
	return true
}
