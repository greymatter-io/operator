package v1alpha1

import "github.com/greymatter-io/operator/pkg/fabric"

func (m Mesh) GenerateEdgeObjects() []fabric.Object {
	var objects []fabric.Object
	// cluster
	// domain
	// listener
	// proxy
	return objects
}

func (m Mesh) GenerateServiceTemplates() fabric.ServiceTemplates {
	return fabric.ServiceTemplates{}
}
