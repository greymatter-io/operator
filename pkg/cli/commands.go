package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/tidwall/gjson"
)

func mkApply(mesh, kind string, data json.RawMessage) cmd {
	key := objKey(kind, data)
	return cmd{
		args:    fmt.Sprintf("apply -t %s -f -", kind),
		requeue: true,
		stdin:   data,
		log: func(out string, err error) {
			if err != nil {
				logger.Error(fmt.Errorf(out), "failed apply", "type", kind, "key", key, "Mesh", mesh)
			} else {
				logger.Info("apply", "type", kind, "key", key, "Mesh", mesh)
			}
		},
		modify: func(out []byte) ([]byte, error) {
			outStr := string(out)
			if !strings.Contains(outStr, "200") {
				return out, fmt.Errorf(string(out))
			}
			return out, nil
		},
	}
}

func mkInjectSVID(mesh, svid, key string) cmd {
	return cmd{
		args:    "get listener --listener-key " + key,
		requeue: true,
		log: func(out string, err error) {
			if err != nil {
				logger.Error(fmt.Errorf(out), "failed get for modify", "type", "listener", "key", key, "Mesh", mesh)
			}
		},
		modify: func(data []byte) ([]byte, error) {
			subjects := gjson.Get(string(data), "secret.subject_names")
			if !subjects.Exists() {
				return nil, fmt.Errorf("listener %s has no secret.subject_names field", key)
			}
			svidMap := make(map[string]struct{})
			for _, subject := range subjects.Array() {
				svidMap[subject.String()] = struct{}{}
			}
			svidMap[fmt.Sprintf("spiffe://greymatter.io/%s", svid)] = struct{}{}
			var svids []string
			for k := range svidMap {
				svids = append(svids, k)
			}
			svidStrs, err := json.Marshal(svids)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal new subject_names from %v", svids)
			}
			patch, err := jsonpatch.DecodePatch([]byte(fmt.Sprintf(`[{"op":"replace","path":"/secret/subject_names","value":%s}]`, svidStrs)))
			if err != nil {
				return nil, fmt.Errorf("failed to generate patch with new subject_names from %v: %w", svids, err)
			}
			modified, err := patch.Apply([]byte(data))
			if err != nil {
				return nil, fmt.Errorf("failed to modify subject names for listener %s from %v: %w", key, svids, err)
			}
			return modified, nil
		},
		then: &cmd{
			args: "apply -t listener -f -",
			log: func(out string, err error) {
				if err != nil {
					logger.Error(fmt.Errorf(out), "failed modify", "type", "listener", "key", key, "Mesh", mesh)
				} else {
					logger.Info("modify", "type", "listener", "key", key, "Mesh", mesh)
				}
			},
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
		log: func(out string, err error) {
			if err != nil {
				logger.Error(fmt.Errorf(out), "failed delete", "type", kind, "key", key, "Mesh", mesh)
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
