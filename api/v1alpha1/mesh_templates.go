package v1alpha1

import "github.com/greymatter-io/operator/pkg/meshobjects"

func (m Mesh) EdgeObjects() []meshobjects.Object {
	var objects []meshobjects.Object
	// cluster
	// domain
	// listener
	// proxy
	return objects
}

func (m Mesh) ServiceTemplates() meshobjects.ServiceTemplates {
	return meshobjects.ServiceTemplates{}
}
