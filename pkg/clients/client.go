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
	mesh        string
	flags       []string
	controlCmds chan cmd
	catalogCmds chan cmd
	f           *fabric.Fabric
}

type cmd struct {
	args  string
	stdin json.RawMessage
	// Attempt to parse stdout into readable key-value pairs, or error
	reader
	// If the cmd (or read) fails, run this cmd
	backup *cmd
}

type result struct {
	kvs []interface{}
	err error
}

type reader func(string) result

func newMeshClient(mesh *v1alpha1.Mesh, flags ...string) *meshClient {
	mc := &meshClient{
		mesh:        mesh.Name,
		flags:       flags,
		controlCmds: make(chan cmd),
		catalogCmds: make(chan cmd),
		f:           fabric.New(mesh),
	}

	// Consume commands to send to Control
	go func(controlCmds chan cmd) {
		// Ping Control every 5s until responsive (getting the Mesh's zone's checksum)
		mc.persist(5, cmd{
			// args: fmt.Sprintf("get zone --zone-key %s", mesh.Spec.Zone),
			args:   fmt.Sprintf("get zone %s", mesh.Spec.Zone),
			reader: pluck("zone_key", "checksum"),
		})
		logger.Info("Connected to Control", "Mesh", mesh.Name)

		// Configure edge objects
		objects := mc.f.Edge()

		for _, c := range []cmd{
			mkApply("domain", objects.Domain),
			mkApply("listener", objects.Listener),
			mkApply("proxy", objects.Proxy),
			mkApply("cluster", objects.Cluster),
		} {
			r := mc.run(c)
			if r.err != nil {
				logger.Error(r.err, c.args, r.kvs...)
			} else {
				logger.Info(c.args, r.kvs...)
			}
		}

		// Then consume additional commands for control objects
		for c := range controlCmds {
			r := mc.run(c)
			if r.err != nil {
				logger.Error(r.err, c.args, r.kvs...)
			} else {
				logger.Info(c.args, r.kvs...)
			}
		}
	}(mc.controlCmds)

	// Consume commands to send to Catalog
	go func(catalogCmds chan cmd) {
		// Ping Catalog every 5s until responsive (getting the Mesh's session status with Control).
		mc.persist(5, cmd{
			// args: fmt.Sprintf("get catalogmesh --mesh-id %s", mesh.Name),
			args:   fmt.Sprintf("get catalog-mesh %s", mesh.Name),
			reader: pluck("mesh_id", "session_statuses.default"),
		})
		logger.Info("Connected to Catalog", "Mesh", mesh.Name)

		// Then consume additional commands for catalog objects
		for c := range catalogCmds {
			r := mc.run(c)
			if r.err != nil {
				logger.Error(r.err, c.args, r.kvs...)
			} else {
				logger.Info(c.args, r.kvs...)
			}
		}
	}(mc.catalogCmds)

	return mc
}

func (mc *meshClient) run(c cmd) result {
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
	parsed := string(out)
	r := result{
		kvs: []interface{}{"output", parsed},
		err: err,
	}
	// If there was no error, and a reader is defined, attempt to read
	if err == nil && c.reader != nil {
		r = c.reader(parsed)
	}
	// If there is an error (either from command or c.reader), and a backup is set, run it
	if err != nil && c.backup != nil {
		return mc.run(*c.backup)
	}
	return r
}

func (mc *meshClient) persist(seconds int, c cmd) bool {
	r := mc.run(c)
	if r.err != nil {
		logger.Error(r.err, c.args, r.kvs...)
		time.Sleep(time.Second * time.Duration(seconds))
		return mc.persist(seconds, c)
	}
	logger.Info(c.args, r.kvs...)

	return true
}

// temp while CLI 4 is being worked on
func mkApply(kind string, data json.RawMessage) cmd {
	kindKey := fmt.Sprintf("%s_key", kind)
	key := pluck(kindKey)(string(data)).kvs[1]
	return cmd{
		args:   fmt.Sprintf("create %s", kind),
		stdin:  data,
		reader: pluck(kindKey, "checksum"),
		backup: &cmd{
			args:   fmt.Sprintf("edit %s %s", kind, key),
			stdin:  data,
			reader: pluck(kindKey, "checksum"),
		},
	}
}

func pluck(keys ...string) reader {
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

func cliVersion() (string, error) {
	r := (&meshClient{}).run(cmd{
		args: "version",
		// args: "--version",
		reader: func(out string) result {
			// split := strings.Split(out, " ")
			// if len(split) < 3 {
			// 	return "", out, fmt.Errorf("failed to get version")
			// }
			// return "version", split[2], nil
			lines := strings.Split(out, "\n")
			if len(lines) != 6 {
				return result{nil, fmt.Errorf("unexpected output")}
			}
			fields := strings.Fields(lines[1])
			if len(fields) != 2 {
				return result{nil, fmt.Errorf("unexpected output")}
			}
			return result{[]interface{}{fields[1]}, nil}
		},
	})
	if len(r.kvs) != 1 {
		return "", r.err
	}
	return r.kvs[0].(string), nil
}
