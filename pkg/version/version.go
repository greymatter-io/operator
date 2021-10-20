package version

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"github.com/go-redis/redis/v8"
	"github.com/greymatter-io/operator/pkg/cueutils"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("version")
)

// A container for a cue.Value that holds all installation configs
// for a release version of Grey Matter, as well as options applied from a Mesh CR.
type Version struct {
	cue cue.Value
}

// Deeply copies the Version's cue.Value into a new Version.
func (v Version) Copy() Version {
	return Version{v.cue}
}

// Implements the functional opts pattern.
func (v *Version) Apply(opts ...InstallOption) {
	for _, opt := range opts {
		opt(v)
	}
}

// An option for mutating the Version's cue.Value.
type InstallOption func(*Version)

func Strings(kvs map[string]string) InstallOption {
	return func(v *Version) {
		var s string
		for k, v := range kvs {
			if v != "" {
				s += fmt.Sprintf(`%s: "%s"`, k, v) + "\n"
			}
		}
		v.cue = v.cue.Unify(cueutils.FromStrings(s))
	}
}

func StringSlices(kvs map[string][]string) InstallOption {
	return func(v *Version) {
		var s string
		for k, v := range kvs {
			if len(v) > 0 {
				var quoted []string
				for _, vv := range v {
					quoted = append(quoted, fmt.Sprintf(`"%s"`, vv))
				}
				s += fmt.Sprintf(`%s: [%s]`, k, strings.Join(quoted, ",")) + "\n"
			}
		}
		v.cue = v.cue.Unify(cueutils.FromStrings(s))
	}
}

func Interfaces(kvs map[string]interface{}) InstallOption {
	return func(v *Version) {
		var s string
		for k, v := range kvs {
			s += fmt.Sprintf(`%s: %v`, k, v) + "\n"
		}
		v.cue = v.cue.Unify(cueutils.FromStrings(s))
	}
}

// An InstallOption for injecting Redis configuration for either an external
// Redis server (if the config is not nil) or otherwise an internal Redis deployment.
func Redis(externalURL string) InstallOption {
	return func(v *Version) {
		// NOTE: Generation happens each time as this option is applied, which will cause a service restart to update envs.
		if externalURL == "" {
			b := make([]byte, 10)
			rand.Read(b)
			password := base64.URLEncoding.EncodeToString(b)
			v.cue = v.cue.Unify(cueutils.FromStrings(fmt.Sprintf(`Redis: password: "%s"`, password)))
			return
		}

		// TODO: In the Mesh validating webhook, ensure the user provided URL is parseable.
		// This actually might be OBE if we require the user to supply values separately rather than as a URL.
		// It makes more sense to do it that way so that the user can store the Redis password in a secret that we reference.
		redisOptions, _ := redis.ParseURL(externalURL)
		hostPort := redisOptions.Addr
		split := strings.Split(hostPort, ":")
		host, port := split[0], split[1]
		password := redisOptions.Password
		db := fmt.Sprintf("%d", redisOptions.DB)
		v.cue = v.cue.Unify(cueutils.FromStrings(fmt.Sprintf(
			`Redis: {
				host: "%s"
				port: "%s"
				password: "%s"
				db: "%s"
			}`,
			host, port, password, db)),
		)
	}
}

// An InstallOption for injecting user tokens to be added to JWT Security.
// Also injects an API key and private key used by the service.
func UserTokens(users string) InstallOption {
	// Assume users is a valid JSON string, since it's been validated by Mesh.InstallOptions().
	var buf bytes.Buffer
	json.Compact(&buf, []byte(users))

	return func(v *Version) {
		v.cue = v.cue.Unify(cueutils.FromStrings(fmt.Sprintf(`
			JWT: userTokens: """
				%s
			"""`, buf.String())),
		)
	}
}

// An InstallOption for injecting generated secret values to be used by JWT Security.
// This may not be needed later on if we can use custom template functions in cueutils.FromStrings (i.e. from Sprig).
// NOTE: Generation happens each time as this option is applied, which will cause a service restart to update envs.
func JWTSecrets(v *Version) {
	// TODO: Generate keys.
	v.cue = v.cue.Unify(cueutils.FromStrings(
		`JWT: {
			apiKey: "MTIzCg=="
			privateKey: "LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1JSGNBZ0VCQkVJQkhRY01yVUh5ZEFFelNnOU1vQWxneFF1a3lqQTROL2laa21ETVIvdFRkVmg3U3hNYk8xVE4KeXdzRkJDdTYvZEZXTE5rUDJGd1FFQmtqREpRZU9mc3hKZWlnQndZRks0RUVBQ09oZ1lrRGdZWUFCQUJEWklJeAp6a082cWpkWmF6ZG1xWFg1dnRFcWtodzlkcVREeTN6d0JkcXBRUmljWDRlS2lZUUQyTTJkVFJtWk0yZE9FRHh1Clhja0hzcVMxZDNtWHBpcDh2UUZHTWJCM1hRVm9DZWN0SUlLMkczRUlwWmhGZFNGdG1sa2t5U1N4angzcS9UcloKaVlRTjhJakpPbUNueUdXZ1VWUkdERURiNWlZdkZXc3dpSkljSWYyOGVRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="
		}`,
	))
}
