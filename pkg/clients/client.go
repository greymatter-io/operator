package clients

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/greymatter-io/operator/api/v1alpha1"
	"github.com/greymatter-io/operator/pkg/fabric"
	"github.com/tidwall/gjson"
)

type meshClient struct {
	mesh  string
	flags []string
	cmds  chan cmd
	f     *fabric.Fabric
}

type cmd struct {
	args string
	read func(string) (string, string, error)
}

func newMeshClient(mesh *v1alpha1.Mesh, flags ...string) (*meshClient, error) {
	f, err := fabric.New(mesh)
	if err != nil {
		return nil, err
	}

	mc := &meshClient{
		mesh:  mesh.Name,
		flags: flags,
		cmds:  make(chan cmd),
		f:     f,
	}

	// Make a channel to notify when we've successfully pinged Control and Catalog.
	pinged := make(chan struct{})

	// Range over pinged until it's closed, and then start listening for additional commands.
	// This goroutine can only be stopped by closing mc.cmds.
	go func(cmds chan cmd, p chan struct{}) {
		<-p
		<-p
		for c := range cmds {
			desc, out, err := mc.run(c)
			if err != nil {
				logger.Error(fmt.Errorf("%s", out), c.args)
			} else {
				logger.Info(c.args, desc, out)
			}
		}
	}(mc.cmds, pinged)

	// Ping Control every 10s until responsive (getting the Mesh's zone's checksum)
	go mc.persist(10, pinged, cmd{
		args: fmt.Sprintf("get zone %s", mesh.Spec.Zone),
		// args: fmt.Sprintf("get zone --zone-key %s", mesh.Spec.Zone),
		read: pluck("checksum"),
	})

	// Ping Catalog every 10s until responsive (getting the Mesh's session status with Control).
	go mc.persist(10, pinged, cmd{
		args: fmt.Sprintf("get catalog-mesh %s", mesh.Name),
		// args: fmt.Sprintf("get catalogmesh --mesh-id %s", mesh.Name),
		read: pluck("session_statuses.default"),
	})

	return mc, nil
}

func (mc *meshClient) run(c cmd) (string, string, error) {
	args := strings.Split(c.args, " ")
	if len(mc.flags) > 0 {
		for _, flag := range mc.flags {
			args = append(strings.Split(flag, " "), args...)
		}
	}
	out, err := exec.Command("greymatter", args...).CombinedOutput()
	if err != nil {
		return "output", string(out), err
	}
	if c.read == nil {
		return "output", string(out), nil
	}
	return c.read(string(out))
}

func (mc *meshClient) persist(seconds int, done chan struct{}, c cmd) {
	desc, out, err := mc.run(c)
	if err != nil {
		logger.Error(fmt.Errorf("%s", out), c.args)
		time.Sleep(time.Second * time.Duration(seconds))
		go mc.persist(seconds, done, c)
	} else {
		logger.Info(c.args, desc, out)
		done <- struct{}{}
	}
}

func pluck(key string) func(string) (string, string, error) {
	return func(out string) (string, string, error) {
		value := gjson.Get(out, key)
		if value.Exists() {
			return key, value.String(), nil
		}
		return "", out, fmt.Errorf("failed to get %s", key)
	}
}

func cliVersion() (string, error) {
	_, version, err := (&meshClient{}).run(cmd{
		args: "--version",
		read: func(out string) (string, string, error) {
			split := strings.Split(out, " ")
			if len(split) < 3 {
				return "", out, fmt.Errorf("failed to get version")
			}
			return "version", split[2], nil
		},
	})
	return version, err
}
