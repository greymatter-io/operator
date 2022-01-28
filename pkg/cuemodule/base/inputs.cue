package base

import (
  "list"
  "strings"
)

Environment: *"kubernetes" | string
MeshName: string
ReleaseVersion: string
Zone: *"default-zone" | string
IngressSubDomain: *"" | string

InstallNamespace: string
WatchNamespaces: [...string]

// Add the install namespace to watch namespaces, and then use list comprehension to identify unique values
allWatchNamespaces: WatchNamespaces + [InstallNamespace]
controlNamespaces: strings.Join([
  for i, ns in allWatchNamespaces if !list.Contains(list.Drop(allWatchNamespaces, i+1), ns) { ns }
], ",")

JWT: {
  userTokens: *"[]" | string
  apiKey: "" | string
  privateKey: "" | string
}
