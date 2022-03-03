package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

func MkApply(kind string, data json.RawMessage) Cmd {
	key := objKey(kind, data)
	return Cmd{
		args:    fmt.Sprintf("apply -t %s -f -", kind),
		requeue: true,
		stdin:   data,
		log: func(out string, err error) {
			if err != nil {
				logger.Error(fmt.Errorf(out), "failed apply", "type", kind, "key", key)
			} else {
				logger.Info("apply", "type", kind, "key", key)
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

func ApplyAll(client *Client, objects []json.RawMessage, kinds []string) {
	for i, kind := range kinds {
		if kind == "catalogservice" { // Catalog is special, because it goes on a different channel
			client.CatalogCmds <- MkApply(kind, objects[i])
		} else if kind != "" { // Everything else goes to Control
			client.ControlCmds <- MkApply(kind, objects[i])
		} else {
			// TODO explode
			logger.Error(nil, "Loaded unexpected object, not recognizable as Grey Matter config", "Object", string(objects[i]))
		}
	}
}

func UnApplyAll(client *Client, objects []json.RawMessage, kinds []string) {
	for i, kind := range kinds {
		if kind == "catalogservice" { // Catalog is special, because it goes on a different channel
			client.CatalogCmds <- mkDelete(kind, objects[i])
		} else if kind != "" { // Everything else goes to Control
			client.ControlCmds <- mkDelete(kind, objects[i])
		} else {
			// TODO explode
			logger.Error(nil, "Loaded unexpected object, not recognizable as Grey Matter config - ignoring", "Object", string(objects[i]))
		}
	}
}

// TODO bring this back when we re-enable SPIRE support so we can inject
// trust between functions
//func mkInjectSVID(mesh, svid, key string) Cmd {
//	return Cmd{
//		args:    "get listener --listener-key " + key,
//		requeue: true,
//		log: func(out string, err error) {
//			if err != nil {
//				logger.Error(fmt.Errorf(out), "failed get for modify", "type", "listener", "key", key, "Mesh", mesh)
//			}
//		},
//		modify: func(data []byte) ([]byte, error) {
//			subjects := gjson.Get(string(data), "secret.subject_names")
//			if !subjects.Exists() {
//				return nil, fmt.Errorf("listener %s has no secret.subject_names field", key)
//			}
//			svidMap := make(map[string]struct{})
//			for _, subject := range subjects.Array() {
//				svidMap[subject.String()] = struct{}{}
//			}
//			svidMap[fmt.Sprintf("spiffe://greymatter.io/%s", svid)] = struct{}{}
//			var svids []string
//			for k := range svidMap {
//				svids = append(svids, k)
//			}
//			svidStrs, err := json.Marshal(svids)
//			if err != nil {
//				return nil, fmt.Errorf("failed to marshal new subject_names from %v", svids)
//			}
//			patch, err := jsonpatch.DecodePatch([]byte(fmt.Sprintf(`[{"op":"replace","path":"/secret/subject_names","value":%s}]`, svidStrs)))
//			if err != nil {
//				return nil, fmt.Errorf("failed to generate patch with new subject_names from %v: %w", svids, err)
//			}
//			modified, err := patch.Apply([]byte(data))
//			if err != nil {
//				return nil, fmt.Errorf("failed to modify subject names for listener %s from %v: %w", key, svids, err)
//			}
//			return modified, nil
//		},
//		then: &Cmd{
//			args: "apply -t listener -f -",
//			log: func(out string, err error) {
//				if err != nil {
//					logger.Error(fmt.Errorf(out), "failed modify", "type", "listener", "key", key, "Mesh", mesh)
//				} else {
//					logger.Info("modify", "type", "listener", "key", key, "Mesh", mesh)
//				}
//			},
//		},
//	}
//}

func mkDelete(kind string, data json.RawMessage) Cmd {
	key := objKey(kind, data)
	args := fmt.Sprintf("delete %s --%s %s", kind, kindFlag(kind), key)
	if kind == "catalogservice" {
		var extracted struct {
			MeshID string `json:"mesh_id"`
		}
		_ = json.Unmarshal(data, &extracted)

		args += fmt.Sprintf(" --mesh-id %s", extracted.MeshID)
	}
	return Cmd{
		args: args,
		log: func(out string, err error) {
			if err != nil {
				logger.Error(fmt.Errorf(out), "failed delete", "type", kind, "key", key)
			} else {
				logger.Info("delete", "type", kind, "key", key)
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
