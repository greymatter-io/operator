package clients

import (
	"bytes"
	"encoding/json"
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
	args  string
	stdin json.RawMessage
	read  func(string) (string, string, error)
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
		logger.Info("Connected to Control and Catalog", "Mesh", mesh.Name)

		// Configure edge objects
		edge := mc.f.Edge()
		mc.apply("domain", edge.Domain)
		mc.apply("listener", edge.Listener)
		mc.apply("proxy", edge.Proxy)
		mc.apply("cluster", edge.Cluster)

		// Then wait for service object commands
		for c := range cmds {
			desc, out, err := mc.run(c)
			if err != nil {
				logger.Error(fmt.Errorf(out), c.args)
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
	command := exec.Command("greymatter", args...)
	if len(c.stdin) > 0 {
		command.Stdin = bytes.NewReader(c.stdin)
	}
	out, err := command.CombinedOutput()
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

// temp while CLI 4 is being worked on
func (mc *meshClient) apply(kind string, data json.RawMessage) {
	c := cmd{
		args:  fmt.Sprintf("create %s", kind),
		stdin: data,
		read:  pluck("checksum"),
	}
	desc, out, err := mc.run(c)
	if err != nil {
		if strings.Contains(out, "duplicate") || strings.Contains(out, "exists") {
			_, objKey, _ := pluck(fmt.Sprintf("%s_key", kind))(string(data))
			c.args = fmt.Sprintf("edit %s %s", kind, objKey)
			desc, out, _ := mc.run(c)
			logger.Info(c.args, desc, out)
			return
		}
		logger.Error(fmt.Errorf(out), c.args)
		return
	}
	logger.Info(c.args, desc, out)
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
		args: "version",
		// args: "--version",
		read: func(out string) (string, string, error) {
			// split := strings.Split(out, " ")
			// if len(split) < 3 {
			// 	return "", out, fmt.Errorf("failed to get version")
			// }
			// return "version", split[2], nil
			lines := strings.Split(out, "\n")
			if len(lines) != 6 {
				return "", out, fmt.Errorf("unexpected output")
			}
			fields := strings.Fields(lines[1])
			if len(fields) != 2 {
				return "", out, fmt.Errorf("unexpected output")
			}
			return "", fields[1], nil
		},
	})
	return version, err
}
