package clients

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
)

type client struct {
	mesh    string
	conf    string
	cmds    chan cmd
	results chan result
	f       *fabric.Fabric
}

func newClient(mesh *v1alpha1.Mesh, conf string) (*client, error) {
	f, err := fabric.New(mesh)
	if err != nil {
		return nil, err
	}

	if conf == "" {
		conf = fmt.Sprintf(`
		[api]
		host = "http://control-api.%s.svc:5555/v1.0"
		[catalog]
		host = "http://catalog.%s.svc:8080"
		mesh = "%s"
		`, mesh.Namespace, mesh.Namespace, mesh.Name)
	}

	conf = base64.StdEncoding.EncodeToString([]byte(conf))

	cl := &client{
		mesh:    mesh.Name,
		conf:    conf,
		cmds:    make(chan cmd),
		results: make(chan result),
		f:       f,
	}

	// Start consuming the client's cmds channel.
	// The channel will close upon cleanup.
	go func(c *client) {
		for cmd := range c.cmds {
			out, err := cmd.run(c.conf)
			if err != nil {
				logger.Error(err, cmd.args, "Mesh", c.mesh)
			}
			c.results <- result{out: out, err: err}
		}
	}(cl)

	// Ping Control and Catalog API
	cl.retry(cmd{args: fmt.Sprintf("get zone %s", mesh.Spec.Zone)})
	logger.Info("Connected to Control", "Mesh", cl.mesh)
	cl.retry(cmd{args: fmt.Sprintf("get catalogmesh %s", mesh.Name)})
	logger.Info("Connected to Catalog", "Mesh", cl.mesh)

	return cl, nil
}

func (cl *client) retry(c cmd) {
	cl.cmds <- c
	for result := range cl.results {
		if result.err != nil {
			logger.Error(result.err, fmt.Sprintf("%s: retrying in 3s", c.args), "Mesh", cl.mesh)
			time.Sleep(time.Second * 3)
			cl.cmds <- c
		} else {
			break
		}
	}
}
