// Package meshobjects defines functions for generating templates for each service in a mesh.
package meshobjects

type Object struct {
	Kind string
	Data string
}

type ServiceTemplates struct {
	Cluster string // deployment
	Route   string // deployment (added to edge domain)
	Proxy   string // deployment
	Locals  map[string]LocalTemplates
}

type LocalTemplates struct {
	Domain   string // deployment:port (added to proxy and listener)
	Listener string // deployment:port (added to proxy)
	Cluster  string // deployment:port (static config localhost:port)
	Route    string // deployment:port (added to local domain)
}
