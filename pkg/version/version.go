package version

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/greymatter-io/operator/pkg/cueutils"

	"cuelang.org/go/cue"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	logger = ctrl.Log.WithName("version")
)

// Version contains a cue.Value that holds all installation templates for a
// version of Grey Matter, plus options applied from a Mesh custom resource.
type Version struct {
	name string
	cue  cue.Value
}

// Copy deep copies a Version's cue.Value into a new Version.
func (v Version) Copy() Version {
	return Version{v.name, v.cue}
}

// Unify gets the lower bound cue.Value of Version.cue and all argument values.
func (v *Version) Unify(ws ...cue.Value) {
	for _, w := range ws {
		v.cue = v.cue.Unify(w)
	}
}

// UserTokens returns a cue.Value for injecting user tokens to be added to JWT Security.
// Also injects an API key and private key used by the service.
func UserTokens(users string) cue.Value {
	// Assume users is a valid JSON string, since it's been validated in the call to Mesh.CueValues().
	var buf bytes.Buffer
	json.Compact(&buf, []byte(users))

	return cueutils.FromStrings(fmt.Sprintf(`
		JWT: userTokens: """
			%s
		"""`, buf.String()),
	)
}

// JWTSecrets returns a cue.Value for injecting generated secret values to be used by JWT Security.
// This may not be needed later on if we can use custom template functions in cueutils.FromStrings (i.e. from Sprig).
// NOTE: Generation happens each time as this option is applied, which will cause a service restart to update envs.
func JWTSecrets() cue.Value {
	// TODO: Generate keys.
	return cueutils.FromStrings(
		`JWT: {
			apiKey: "MTIzCg=="
			privateKey: "LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1JSGNBZ0VCQkVJQkhRY01yVUh5ZEFFelNnOU1vQWxneFF1a3lqQTROL2laa21ETVIvdFRkVmg3U3hNYk8xVE4KeXdzRkJDdTYvZEZXTE5rUDJGd1FFQmtqREpRZU9mc3hKZWlnQndZRks0RUVBQ09oZ1lrRGdZWUFCQUJEWklJeAp6a082cWpkWmF6ZG1xWFg1dnRFcWtodzlkcVREeTN6d0JkcXBRUmljWDRlS2lZUUQyTTJkVFJtWk0yZE9FRHh1Clhja0hzcVMxZDNtWHBpcDh2UUZHTWJCM1hRVm9DZWN0SUlLMkczRUlwWmhGZFNGdG1sa2t5U1N4angzcS9UcloKaVlRTjhJakpPbUNueUdXZ1VWUkdERURiNWlZdkZXc3dpSkljSWYyOGVRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo="
		}`,
	)
}
