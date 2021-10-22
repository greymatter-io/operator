package base

import (
  "list"
  "strings"
)

MeshName: string

// Where to install components
InstallNamespace: string

Zone: *"default-zone" | string

IngressSubDomain: *"localhost" | string

// The scope of the mesh network
WatchNamespaces: [...string]

// Add the install namespace to watch namespaces, and then use list comprehension to identify unique values
allWatchNamespaces: WatchNamespaces + [InstallNamespace]
controlNamespaces: strings.Join([
  for i, ns in allWatchNamespaces if !list.Contains(list.Drop(allWatchNamespaces, i+1), ns) { ns }
], ",")

Spire: *false | bool

JWT: {
  userTokens: *"[]" | string
  apiKey: "" | string
  privateKey: "" | string
}

Redis: {
  host: *"gm-redis.\(InstallNamespace).svc.cluster.local" | string
  port: *"6379" | string
  password: "" | string
  db: *"0" | string
}
