package meshobjects

import (
	"fmt"
	"sync"

	"github.com/bcmendoza/gm-operator/controllers/gmcore"
)

type Cache struct {
	sync.RWMutex
	revisions map[Revision]struct{}
}

func NewCache() *Cache {
	return &Cache{revisions: make(map[Revision]struct{})}
}

func (c *Cache) Missing(mesh string) map[Revision]struct{} {
	missing := make(map[Revision]struct{})

	// todo: initialize at startup
	var revisions []Revision
	revisions = append(revisions, Revision{Mesh: mesh, Kind: "Zone", Key: mesh})
	revisions = append(revisions, mkSidecarRevisions(mesh, "edge")...)
	revisions = append(revisions, mkDefaultRevisions(mesh, string(gmcore.ControlApi))...)
	revisions = append(revisions, mkDefaultRevisions(mesh, string(gmcore.Dashboard))...)
	revisions = append(revisions, mkDefaultRevisions(mesh, string(gmcore.JwtSecurity))...)

	c.RLock()
	for _, rev := range revisions {
		if _, ok := c.revisions[rev]; !ok {
			missing[rev] = struct{}{}
		}
	}
	c.RUnlock()

	return missing
}

func (c *Cache) AddZone(mesh string) {
	c.Lock()
	defer c.Unlock()

	c.revisions[Revision{Mesh: mesh, Kind: "Zone", Key: mesh}] = struct{}{}
}

func (c *Cache) AddSidecar(mesh, svcName string) {
	c.Lock()
	defer c.Unlock()

	for _, revision := range mkSidecarRevisions(mesh, svcName) {
		c.revisions[revision] = struct{}{}
	}
}

func (c *Cache) AddService(mesh, svcName string) {
	c.Lock()
	defer c.Unlock()

	for _, revision := range mkDefaultRevisions(mesh, svcName) {
		c.revisions[revision] = struct{}{}
	}
}

func mkSidecarRevisions(mesh, svcName string) []Revision {
	return []Revision{
		{Mesh: mesh, Kind: "Domain", Key: fmt.Sprintf("%s.%s", mesh, svcName)},
		{Mesh: mesh, Kind: "Listener", Key: fmt.Sprintf("%s.%s", mesh, svcName)},
		{Mesh: mesh, Kind: "Proxy", Key: fmt.Sprintf("%s.%s", mesh, svcName)},
		{Mesh: mesh, Kind: "Cluster", Key: fmt.Sprintf("%s.%s", mesh, svcName)},
	}
}

func mkDefaultRevisions(mesh, svcName string) []Revision {
	var revisions []Revision
	revisions = append(revisions, mkSidecarRevisions(mesh, svcName)...)
	revisions = append(revisions,
		Revision{Mesh: mesh, Kind: "Cluster", Key: fmt.Sprintf("%s.%s.service", mesh, svcName)},
		Revision{Mesh: mesh, Kind: "Route", Key: fmt.Sprintf("%s.%s.a", mesh, svcName)},
		Revision{Mesh: mesh, Kind: "Route", Key: fmt.Sprintf("%s.%s.b", mesh, svcName)},
		Revision{Mesh: mesh, Kind: "Route", Key: fmt.Sprintf("%s.%s.c", mesh, svcName)},
	)
	return revisions
}

type Revision struct {
	Mesh string
	Kind string
	Key  string
}
