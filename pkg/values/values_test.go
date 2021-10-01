package values

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"github.com/kylelemons/godebug/diff"
)

//go:embed base.cue
var baseCue string

//go:embed fixture.cue
var fixtureCue string

// Tests the Cue schema defined in base.cue.
// Fails if its expressions cannot be composed into a valid cue.Value.
// A passing schema is suitable for validating our versions files.
func TestBaseSchema(t *testing.T) {
	buildCueValueOrDie(t, baseCue)
}

// Tests fixture.cue against base.cue.
func TestCUEFixture(t *testing.T) {
	buildCueValueOrDie(t, baseCue, fixtureCue)
}

// Tests the applied EdgeValuesFromProxy opt against base.cue+inline.
// TODO: Replace fixture with each versions file.
// func TestEdgeValuesFromProxy(t *testing.T) {
// 	v := &Values{}
// 	yaml.Unmarshal([]byte(fixture), v)
// 	v.Apply(EdgeValuesFromProxy)
// 	validate(t, v, `
// 		proxy: {}
// 		edge: {
// 			image: proxy.image
// 			envs: proxy.envs & {
// 				"XDS_CLUSTER": "edge"
// 			}
// 		}
// 	`)
// }

// Tests the applied SPIRE opt against base.cue+inline.
// TODO: Replace fixture with each versions file.
// func TestSPIRE(t *testing.T) {
// 	v := &Values{}
// 	yaml.Unmarshal([]byte(fixture), v)
// 	v.Apply(SPIRE)
// 	validate(t, v, `
// 		proxy: {
// 			volumes: {
// 				"spire-socket": {
// 					hostPath: {
// 						path: "/run/spire/socket"
// 						type: "DirectoryOrCreate"
// 					}
// 				}
// 			}
// 			volumeMounts: {
// 				"spire-socket": {
// 					mountPath: "/run/spire/socket"
// 				}
// 			}
// 			envs: {
// 				"SPIRE_PATH": "/run/spire/socket/agent.sock"
// 			}
// 		}
// 	`)
// }

// Tests the applied Redis opt for internal Redis against base.cue+inline.
// TODO: Replace fixture with each versions file.
// func TestRedisInternal(t *testing.T) {
// 	v := &Values{}
// 	yaml.Unmarshal([]byte(fixture), v)
// 	v.Apply(Redis(nil, "namespace"))
// 	validate(t, v, fmt.Sprintf(`
// 		import "strings"

// 		control_api: {
// 			envs: {
// 				"GM_CONTROL_API_REDIS_HOST": "greymatter-redis.namespace.svc.cluster.local"
// 				"GM_CONTROL_API_REDIS_PORT": "6379"
// 				"GM_CONTROL_API_REDIS_PASS": redis.envs.REDIS_PASSWORD
// 				"GM_CONTROL_API_REDIS_DB": "0"
// 			}
// 		}
// 		catalog: {
// 			envs: {
// 				"GM_CONTROL_API_REDIS_HOST": "greymatter-redis.namespace.svc.cluster.local"
// 				"GM_CONTROL_API_REDIS_PORT": "6379"
// 				"GM_CONTROL_API_REDIS_PASS": redis.envs.REDIS_PASSWORD
// 				"GM_CONTROL_API_REDIS_DB": "0"
// 			}
// 		}
// 		jwt_security: {
// 			envs: {
// 				"GM_CONTROL_API_REDIS_HOST": "greymatter-redis.namespace.svc.cluster.local"
// 				"GM_CONTROL_API_REDIS_PORT": "6379"
// 				"GM_CONTROL_API_REDIS_PASS": redis.envs.REDIS_PASSWORD
// 				"GM_CONTROL_API_REDIS_DB": "0"
// 			}
// 		}
// 		redis: {
// 			envs: {
// 				"REDIS_PASSWORD": =~ "^.{16}$" & "%s"
// 			}
// 		}
// 	`, v.Redis.Envs["REDIS_PASSWORD"]))
// }

// Tests the applied Redis opt for external Redis against base.cue+inline.
// TODO: Replace fixture with each versions file.
// func TestRedisExternal(t *testing.T) {
// 	v := &Values{}
// 	yaml.Unmarshal([]byte(fixture), v)
// 	cfg := &ExternalRedisConfig{URL: "redis://:mypass@127.0.0.1:6379"}
// 	v.Apply(Redis(cfg, ""))
// 	validate(t, v, `
// 		control_api: {
// 			envs: {
// 				"GM_CONTROL_API_REDIS_HOST": "127.0.0.1"
// 				"GM_CONTROL_API_REDIS_PORT": "6379"
// 				"GM_CONTROL_API_REDIS_PASS": "mypass"
// 				"GM_CONTROL_API_REDIS_DB": "0"
// 			}
// 		}
// 		catalog: {
// 			envs: {
// 				"GM_CONTROL_API_REDIS_HOST": "127.0.0.1"
// 				"GM_CONTROL_API_REDIS_PORT": "6379"
// 				"GM_CONTROL_API_REDIS_PASS": "mypass"
// 				"GM_CONTROL_API_REDIS_DB": "0"
// 			}
// 		}
// 		jwt_security: {
// 			envs: {
// 				"GM_CONTROL_API_REDIS_HOST": "127.0.0.1"
// 				"GM_CONTROL_API_REDIS_PORT": "6379"
// 				"GM_CONTROL_API_REDIS_PASS": "mypass"
// 				"GM_CONTROL_API_REDIS_DB": "0"
// 			}
// 		}
// 	`)
// }

// func validate(t *testing.T, v *Values, inlines ...string) {
// 	schema := buildSchemaOrDie(t, true, inlines...)
// 	valuesYAML, err := yaml.Marshal(v)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	validateYAML(t, schema, valuesYAML)
// }

// func validateYAML(t *testing.T, schema cue.Value, data []byte, inlines ...string) {
// 	if err := cueyaml.Validate(data, schema); err != nil {
// 		for _, e := range errors.Errors(err) {
// 			t.Error(e)
// 		}
// 	}
// 	valuesCue := buildSchemaFromYAMLOrDie(t, data)
// 	printCueDiff(schema, valuesCue)
// }

func buildCueValueOrDie(t *testing.T, inlines ...string) cue.Value {
	if len(inlines) == 0 {
		t.Fatal("no inline Cue build args passed")
	}

	cwd, _ := os.Getwd()
	abs := func(path string) string {
		return filepath.Join(cwd, path)
	}

	ctx := cuecontext.New()
	var values []cue.Value
	for idx, inline := range inlines {
		mockFile := fmt.Sprintf("./mock/%d.cue", idx)
		cfg := &load.Config{Overlay: map[string]load.Source{
			abs(mockFile): load.FromString(inline),
		}}
		for _, i := range load.Instances([]string{mockFile}, cfg) {
			if i.Err != nil {
				t.Fatal(i.Err)
			}
			value := ctx.BuildInstance(i)
			if err := value.Err(); err != nil {
				t.Fatal(err)
			}
			values = append(values, value)
		}
	}

	value := values[0]
	if len(values) > 1 {
		for i := 1; i < len(values); i++ {
			if i == len(values)-1 {
				printCueDiff(value, values[i])
			}
			value = value.Unify(values[i])
			if err := value.Err(); err != nil {
				t.Fatal(err)
			}
		}
	}

	if err := value.Validate(); err != nil {
		for _, e := range errors.Errors(err) {
			t.Error(e)
		}
	}

	return value
}

func printCueDiff(a, b cue.Value) {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	aLines := strings.Split(aStr, "\n")
	bLines := strings.Split(bStr, "\n")

	chunks := diff.DiffChunks(aLines, bLines)

	buf := new(bytes.Buffer)
	for _, c := range chunks {
		for _, line := range c.Added {
			fmt.Fprintf(buf, "\033[32m+%s\n", line)
		}
		for _, line := range c.Deleted {
			fmt.Fprintf(buf, "\033[31m-%s\n", line)
		}
	}
	fmt.Println(strings.TrimRight(buf.String(), "\n"))
}

// func buildSchemaFromYAMLOrDie(t *testing.T, data []byte) cue.Value {
// 	f, err := cueyaml.Extract("", data)
// 	if err != nil {
// 		t.Fatal(f)
// 	}
// 	return cuecontext.New().BuildFile(f)
// }
