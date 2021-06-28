package meshobjects

import (
	"fmt"
	"sync"

	"github.com/go-logr/logr"

	"github.com/greymatter.io/operator/pkg/gmcore"
)

type Cache struct {
	sync.RWMutex
	revisions map[Revision]string
	meshes    map[string][]Revision
}

type Revision struct {
	Mesh string
	Kind string
	Key  string
}

func NewCache() *Cache {
	return &Cache{
		revisions: make(map[Revision]string),
		meshes:    make(map[string][]Revision),
	}
}

func (c *Cache) Register(mesh string, logger logr.Logger) {
	c.RLock()
	_, ok := c.meshes[mesh]
	c.RUnlock()
	if ok {
		return
	}

	logger.Info("Registering mesh '" + mesh + "' with Control API")

	var revisions []Revision
	revisions = append(revisions, mkSidecarRevisions(mesh, "edge")...)

	for _, svcName := range []string{
		string(gmcore.ControlApi),
		string(gmcore.Catalog),
		string(gmcore.JwtSecurity),
		string(gmcore.Dashboard),
	} {
		revisions = append(revisions, mkRevisions(mesh, svcName)...)
	}

	c.Lock()
	c.meshes[mesh] = revisions
	c.Unlock()
}

func (c *Cache) Deregister(mesh string, logger logr.Logger) {
	c.Lock()
	defer c.Unlock()

	if _, ok := c.meshes[mesh]; !ok {
		return
	}

	for revision, checksum := range c.revisions {
		if revision.Mesh == mesh {
			delete(c.revisions, revision)
			logger.Info(
				"Unconfigured "+revision.Kind,
				revision.Kind+"Key", revision.Key,
				"Checksum", checksum,
			)
		}
	}

	delete(c.meshes, mesh)

	logger.Info(fmt.Sprintf("Deregistered mesh '%s' from Control API", mesh))
}

func (c *Cache) Add(revision Revision, checksum string) {
	c.Lock()
	defer c.Unlock()

	c.revisions[revision] = checksum
}

func (c *Cache) Has(revision Revision) bool {
	c.RLock()
	defer c.RUnlock()

	_, ok := c.revisions[revision]
	return ok
}

func (c *Cache) Diff(mesh string) map[Revision]struct{} {
	missing := make(map[Revision]struct{})

	c.RLock()
	revisions, ok := c.meshes[mesh]
	c.RUnlock()
	if !ok {
		return missing
	}

	for _, rev := range revisions {
		if _, ok := c.revisions[rev]; !ok {
			missing[rev] = struct{}{}
		}
	}

	return missing
}

func mkRevisions(mesh, svcName string) []Revision {
	var revisions []Revision
	revisions = append(revisions, mkSidecarRevisions(mesh, svcName)...)
	revisions = append(revisions, mkServiceRevisions(mesh, svcName)...)
	return revisions
}

func mkSidecarRevisions(mesh, svcName string) []Revision {
	return []Revision{
		{Mesh: mesh, Kind: "Domain", Key: fmt.Sprintf("%s.%s", mesh, svcName)},
		{Mesh: mesh, Kind: "Listener", Key: fmt.Sprintf("%s.%s", mesh, svcName)},
		{Mesh: mesh, Kind: "Proxy", Key: fmt.Sprintf("%s.%s", mesh, svcName)},
		{Mesh: mesh, Kind: "Cluster", Key: fmt.Sprintf("%s.%s", mesh, svcName)},
	}
}

func mkServiceRevisions(mesh, svcName string) []Revision {
	revisions := []Revision{
		{Mesh: mesh, Kind: "Cluster", Key: fmt.Sprintf("%s.%s.service", mesh, svcName)},
		{Mesh: mesh, Kind: "Route", Key: fmt.Sprintf("%s.%s.a", mesh, svcName)},
		{Mesh: mesh, Kind: "Route", Key: fmt.Sprintf("%s.%s.b", mesh, svcName)},
	}

	if svcName != "dashboard" {
		revisions = append(
			revisions,
			Revision{Mesh: mesh, Kind: "Route", Key: fmt.Sprintf("%s.%s.c", mesh, svcName)},
		)
	}
	return revisions
}
