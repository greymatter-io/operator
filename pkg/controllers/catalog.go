package controllers

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	v1 "github.com/greymatter-io/operator/pkg/api/v1"
	"github.com/greymatter-io/operator/pkg/clients/catalog"
)

// reconcileMesh reconciles entries in Catalog until all expected objects exist
// This should be called in a goroutine.
func reconcileCatalog(controller *MeshController, mesh *v1.Mesh, logger logr.Logger) {
	addr := fmt.Sprintf("http://catalog.%s.svc:9080", mesh.Namespace)
	catalog := catalog.NewClient(mesh.Spec.Version, addr, logger)

	// If any of the operations fail, retry
	if !catalog.CreateMesh(mesh.Name, mesh.Namespace) {
		time.Sleep(time.Second * 2)
		go reconcileCatalog(controller, mesh, logger)
		return
	}

	if !catalog.CreateService(
		"control-api",
		mesh.Name,
		"Grey Matter Control API",
		"latest",
		"The purpose of the Grey Matter Control API is to update the configuration of every Grey Matter Proxy in the mesh.",
		"Decipher",
		"services/control-api/latest/v1.0",
		"/services/control-api/latest/",
		"Core Mesh") {
		time.Sleep(time.Second * 2)
		go reconcileCatalog(controller, mesh, logger)
		return
	}

	if !catalog.CreateService(
		"catalog",
		mesh.Name,
		"Grey Matter Catalog",
		"latest",
		"The Grey Matter Catalog service interfaces with the Fabric mesh xDS interface to provide high level summaries and more easily consumable views of the current state of the mesh. It powers the Grey Matter application and any other applications that need to understand what is present in the mesh.",
		"Decipher",
		"services/catalog/latest/",
		"/services/catalog/latest/",
		"Core Mesh") {
		time.Sleep(time.Second * 2)
		go reconcileCatalog(controller, mesh, logger)
		return
	}

	if !catalog.CreateService(
		"dashboard",
		mesh.Name,
		"Grey Matter Dashboard",
		"latest",
		"The Grey Matter application is a user dashboard that paints a high-level picture of the service mesh.",
		"Decipher",
		"services/dashboard/latest/",
		"/services/dashboard/latest/",
		"Core Mesh") {
		time.Sleep(time.Second * 2)
		go reconcileCatalog(controller, mesh, logger)
		return
	}

	if !catalog.CreateService(
		"jwt-security",
		mesh.Name,
		"Grey Matter JWT Security",
		"latest",
		"The Grey Matter JWT security service is a JWT Token generation and retrieval service.",
		"Decipher",
		"services/jwt-security/latest/",
		"/services/jwt-security/latest/",
		"Core Mesh") {
		time.Sleep(time.Second * 2)
		go reconcileCatalog(controller, mesh, logger)
		return
	}

	// todo: if mesh.Spec.Version == "1.3", create SLO service card
}
