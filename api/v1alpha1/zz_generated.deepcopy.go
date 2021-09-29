// +build !ignore_autogenerated

/*
Copyright Decipher Technology Studios 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExternalRedisConfig) DeepCopyInto(out *ExternalRedisConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExternalRedisConfig.
func (in *ExternalRedisConfig) DeepCopy() *ExternalRedisConfig {
	if in == nil {
		return nil
	}
	out := new(ExternalRedisConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstallValues) DeepCopyInto(out *InstallValues) {
	*out = *in
	if in.Proxy != nil {
		in, out := &in.Proxy, &out.Proxy
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
	if in.Edge != nil {
		in, out := &in.Edge, &out.Edge
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
	if in.Control != nil {
		in, out := &in.Control, &out.Control
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
	if in.ControlAPI != nil {
		in, out := &in.ControlAPI, &out.ControlAPI
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
	if in.Catalog != nil {
		in, out := &in.Catalog, &out.Catalog
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
	if in.Dashboard != nil {
		in, out := &in.Dashboard, &out.Dashboard
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
	if in.JWTSecurity != nil {
		in, out := &in.JWTSecurity, &out.JWTSecurity
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
	if in.Redis != nil {
		in, out := &in.Redis, &out.Redis
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
	if in.Prometheus != nil {
		in, out := &in.Prometheus, &out.Prometheus
		*out = new(Values)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstallValues.
func (in *InstallValues) DeepCopy() *InstallValues {
	if in == nil {
		return nil
	}
	out := new(InstallValues)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstallationConfig) DeepCopyInto(out *InstallationConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.InstallValues.DeepCopyInto(&out.InstallValues)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstallationConfig.
func (in *InstallationConfig) DeepCopy() *InstallationConfig {
	if in == nil {
		return nil
	}
	out := new(InstallationConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *InstallationConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *InstallationConfigList) DeepCopyInto(out *InstallationConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]InstallationConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new InstallationConfigList.
func (in *InstallationConfigList) DeepCopy() *InstallationConfigList {
	if in == nil {
		return nil
	}
	out := new(InstallationConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *InstallationConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Mesh) DeepCopyInto(out *Mesh) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Mesh.
func (in *Mesh) DeepCopy() *Mesh {
	if in == nil {
		return nil
	}
	out := new(Mesh)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Mesh) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshList) DeepCopyInto(out *MeshList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Mesh, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshList.
func (in *MeshList) DeepCopy() *MeshList {
	if in == nil {
		return nil
	}
	out := new(MeshList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *MeshList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshSpec) DeepCopyInto(out *MeshSpec) {
	*out = *in
	if in.ExternalRedis != nil {
		in, out := &in.ExternalRedis, &out.ExternalRedis
		*out = new(ExternalRedisConfig)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshSpec.
func (in *MeshSpec) DeepCopy() *MeshSpec {
	if in == nil {
		return nil
	}
	out := new(MeshSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MeshStatus) DeepCopyInto(out *MeshStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MeshStatus.
func (in *MeshStatus) DeepCopy() *MeshStatus {
	if in == nil {
		return nil
	}
	out := new(MeshStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedisConfig) DeepCopyInto(out *RedisConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedisConfig.
func (in *RedisConfig) DeepCopy() *RedisConfig {
	if in == nil {
		return nil
	}
	out := new(RedisConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceGroup) DeepCopyInto(out *ResourceGroup) {
	*out = *in
	if in.Deployment != nil {
		in, out := &in.Deployment, &out.Deployment
		*out = new(v1.Deployment)
		(*in).DeepCopyInto(*out)
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make([]*corev1.Service, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(corev1.Service)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceGroup.
func (in *ResourceGroup) DeepCopy() *ResourceGroup {
	if in == nil {
		return nil
	}
	out := new(ResourceGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Values) DeepCopyInto(out *Values) {
	*out = *in
	if in.Arguments != nil {
		in, out := &in.Arguments, &out.Arguments
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make(map[string]corev1.ContainerPort, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Envs != nil {
		in, out := &in.Envs, &out.Envs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.EnvsFrom != nil {
		in, out := &in.EnvsFrom, &out.EnvsFrom
		*out = make(map[string]corev1.EnvVarSource, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make(map[string]corev1.VolumeSource, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.VolumeMounts != nil {
		in, out := &in.VolumeMounts, &out.VolumeMounts
		*out = make(map[string]corev1.VolumeMount, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Values.
func (in *Values) DeepCopy() *Values {
	if in == nil {
		return nil
	}
	out := new(Values)
	in.DeepCopyInto(out)
	return out
}
