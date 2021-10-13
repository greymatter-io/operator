package clients

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/greymatter-io/operator/pkg/fabric"
)

type client struct {
	cmds chan meshCmd
	f    *fabric.Fabric
}

func newClient(zone string, meshPort int32) (*client, error) {
	f, err := fabric.New(zone, meshPort)
	if err != nil {
		return nil, err
	}

	c := &client{
		cmds: make(chan meshCmd),
		f:    f,
	}

	// Start consuming the client's cmds channel.
	// The channel will close upon cleanup.
	go func(cl *client) {
		for cmd := range c.cmds {
			cmd.run()
		}
	}(c)

	// Ping Control API and Catalog
	c.cmds <- meshCmd{} // TODO: ping control api
	c.cmds <- meshCmd{} // TODO: ping catalog api

	return c, nil
}

type meshCmd struct {
	args string
	read func(string) (string, error)
}

func (mc meshCmd) run() (string, error) {
	cmd := exec.Command("greymatter", strings.Split(mc.args, " ")...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	if mc.read == nil {
		return out.String(), nil
	}
	result, err := mc.read(out.String())
	if err != nil {
		return "", fmt.Errorf("failed to format: %w", err)
	}
	return result, nil
}

func cliVersion() (string, error) {
	return meshCmd{
		args: "version",
		read: func(out string) (string, error) {
			lines := strings.Split(out, "\n")
			if len(lines) != 6 {
				return "", fmt.Errorf("unexpected output: %s", out)
			}
			fields := strings.Fields(lines[1])
			if len(fields) != 2 {
				return "", fmt.Errorf("unexpected output: %s", out)
			}
			return fields[1], nil
		},
	}.run()
}
