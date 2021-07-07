package controllers

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	v1 "github.com/greymatter.io/operator/pkg/api/v1"
	"github.com/greymatter.io/operator/pkg/gmcore"
	"github.com/greymatter.io/operator/pkg/meshobjects"
)

// reconcileMesh reconciles mesh objects in Control API until all expected objects exist
// This should be called in a goroutine.
func reconcileMesh(controller *MeshController, mesh *v1.Mesh, logger logr.Logger) {
	addr := fmt.Sprintf("http://control-api.%s.svc:5555", mesh.Namespace)
	api := meshobjects.NewClient(addr, logger)

	// If any of the operations fail, retry
	if err := api.MkProxy(mesh.Name, "edge"); err != nil {
		logger.Error(err, "failed to make edge proxy meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}
	if err := api.MkProxy(mesh.Name, string(gmcore.ControlApi)); err != nil {
		logger.Error(err, "failed to make control-api proxy meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}
	if err := api.MkService(mesh.Name, string(gmcore.ControlApi), "5555"); err != nil {
		logger.Error(err, "failed to make control-api service meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}
	if err := api.MkProxy(mesh.Name, string(gmcore.Catalog)); err != nil {
		logger.Error(err, "failed to make catalog proxy meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}
	if err := api.MkService(mesh.Name, string(gmcore.Catalog), "9080"); err != nil {
		logger.Error(err, "failed to make catalog service meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}
	if err := api.MkProxy(mesh.Name, string(gmcore.Dashboard)); err != nil {
		logger.Error(err, "failed to make dashboard proxy meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}
	if err := api.MkService(mesh.Name, string(gmcore.Dashboard), "1337"); err != nil {
		logger.Error(err, "failed to make dashboard service meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}
	if err := api.MkProxy(mesh.Name, string(gmcore.JwtSecurity)); err != nil {
		logger.Error(err, "failed to make jwt-security proxy meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}
	if err := api.MkService(mesh.Name, string(gmcore.JwtSecurity), "3000"); err != nil {
		logger.Error(err, "failed to make jwt-security service meshobjects")
		time.Sleep(time.Second * 3)
		go reconcileMesh(controller, mesh, logger)
		return
	}

	if mesh.Spec.Version == "1.3" {
		if err := api.MkProxy(mesh.Name, string(gmcore.Slo)); err != nil {
			logger.Error(err, "failed to make slo proxy meshobjects")
			time.Sleep(time.Second * 3)
			go reconcileMesh(controller, mesh, logger)
			return
		}
		if err := api.MkService(mesh.Name, string(gmcore.Slo), "9080"); err != nil {
			logger.Error(err, "failed to make slo service meshobjects")
			time.Sleep(time.Second * 3)
			go reconcileMesh(controller, mesh, logger)
			return
		}
	}
}
