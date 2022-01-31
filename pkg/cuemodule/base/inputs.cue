package base

import (
  "list"
  "strings"
  "encoding/json"

  "github.com/greymatter-io/operator/api/v1alpha1"
)

// A Mesh CR being applied.
mesh: v1alpha1.#Mesh

Environment: *"kubernetes" | mesh.spec.environment
MeshName: mesh.metadata.name
ReleaseVersion: mesh.spec.release_version
Zone: *"default-zone" | mesh.spec.zone

IngressSubDomain: *"" | string

InstallNamespace: mesh.spec.install_namespace
WatchNamespaces: mesh.spec.watch_namespaces

// Secrets that can live in other namespaces so users aren't tied to
// docker.greymatter.io
PullSecrets: mesh.spec.pull_secrets

// Add the install namespace to watch namespaces, and then use list comprehension to identify unique values
allWatchNamespaces: WatchNamespaces + [InstallNamespace]
controlNamespaces: strings.Join([
  for i, ns in allWatchNamespaces if !list.Contains(list.Drop(allWatchNamespaces, i+1), ns) { ns }
], ",")

JWT: {
  apiKey: *"MTIzCg==" | string
  privateKey: *"LS0tLS1CRUdJTiBFQyBQUklWQVRFIEtFWS0tLS0tCk1JSGNBZ0VCQkVJQkhRY01yVUh5ZEFFelNnOU1vQWxneFF1a3lqQTROL2laa21ETVIvdFRkVmg3U3hNYk8xVE4KeXdzRkJDdTYvZEZXTE5rUDJGd1FFQmtqREpRZU9mc3hKZWlnQndZRks0RUVBQ09oZ1lrRGdZWUFCQUJEWklJeAp6a082cWpkWmF6ZG1xWFg1dnRFcWtodzlkcVREeTN6d0JkcXBRUmljWDRlS2lZUUQyTTJkVFJtWk0yZE9FRHh1Clhja0hzcVMxZDNtWHBpcDh2UUZHTWJCM1hRVm9DZWN0SUlLMkczRUlwWmhGZFNGdG1sa2t5U1N4angzcS9UcloKaVlRTjhJakpPbUNueUdXZ1VWUkdERURiNWlZdkZXc3dpSkljSWYyOGVRPT0KLS0tLS1FTkQgRUMgUFJJVkFURSBLRVktLS0tLQo=" | string
  userTokens: *"[]" | string
  if mesh.spec.user_tokens != _|_ {
		userTokens: json.Marshal(mesh.spec.user_tokens)
	}
}
