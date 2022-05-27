// K8s manifests necessary for spire and openshift

package only

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

openshift_privileged_scc: [
	rbacv1.#ClusterRole & {
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "ClusterRole"
		metadata: {
			annotations: {
				"include.release.openshift.io/self-managed-high-availability": "true"
				"rbac.authorization.kubernetes.io/autoupdate":                 "true"
			}
			name: "system:openshift:scc:gm-operator"
		}
		rules: [{
			apiGroups: ["security.openshift.io"]
			resourceNames: ["privileged"]
			resources: ["securitycontextconstraints"]
			verbs: ["use"]
		}]
	},
	rbacv1.#ClusterRoleBinding & {
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "ClusterRoleBinding"
		metadata: {
			name: "system:openshift:scc:gm-operator"
		}
		roleRef: {
			apiGroup: "rbac.authorization.k8s.io"
			kind:     "ClusterRole"
			name:     "system:openshift:scc:gm-operator"
		}
		subjects: [{
			kind:      "ServiceAccount"
			name:      "gm-operator"
			namespace: "gm-operator"
		}]
	},
]

openshift_spire_scc: [
	{// SCC https://spiffe.io/docs/latest/deploying/spire_agent/#security-context-constraints
		allowHostDirVolumePlugin: true
		allowHostIPC:             true
		allowHostNetwork:         true
		allowHostPID:             true
		allowHostPorts:           true
		allowPrivilegeEscalation: true
		allowPrivilegedContainer: false
		allowedCapabilities:      null
		allowedUnsafeSysctls:     null
		apiVersion:               "security.openshift.io/v1"
		defaultAddCapabilities:   null
		fsGroup: type: "MustRunAs"
		groups: []
		kind: "SecurityContextConstraints"
		metadata: {
			annotations: {
				"include.release.openshift.io/self-managed-high-availability": "true"
				"kubernetes.io/description":                                   "Customized policy for Spire to enable host level access."
				"release.openshift.io/create-only":                            "true"
			}
			name: "spire"
		}
		priority:               null
		readOnlyRootFilesystem: false
		requiredDropCapabilities: [ "KILL", "MKNOD", "SETUID", "SETGID"]
		runAsUser: type:          "RunAsAny"
		seLinuxContext: type:     "MustRunAs"
		supplementalGroups: type: "RunAsAny"
		users: []
		volumes: [
			"hostPath",
			"configMap",
			"downwardAPI",
			"emptyDir",
			"persistentVolumeClaim",
			"projected",
			"secret",
		]
	},
	rbacv1.#ClusterRole & {// spire SCC cluster role bound by everything that needs the spire SCC
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "ClusterRole"
		metadata: {
			annotations: {
				"include.release.openshift.io/self-managed-high-availability": "true"
				"rbac.authorization.kubernetes.io/autoupdate":                 "true"
			}
			name: "system:openshift:scc:spire"
		}
		rules: [{
			apiGroups: ["security.openshift.io"]
			resourceNames: ["spire"]
			resources: ["securitycontextconstraints"]
			verbs: ["use"]
		}]
	},
	rbacv1.#ClusterRoleBinding & {// The operator's own binding to the spire scc, so it can grant it to other things
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "ClusterRoleBinding"
		metadata: {
			name: "system:openshift:scc:spire:gm-operator"
		}
		roleRef: {
			apiGroup: "rbac.authorization.k8s.io"
			kind:     "ClusterRole"
			name:     "system:openshift:scc:spire"
		}
		subjects: [{
			kind:      "ServiceAccount"
			name:      "gm-operator"
			namespace: "gm-operator"
		}]
	},
]

openshift_spire: [
	// The actual ClusterRole these refer to is applied at install-time if openshift and spire are both on - see above
	rbacv1.#RoleBinding & {
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "RoleBinding"
		metadata: {
			name:      "system:openshift:scc:spire:agent"
			namespace: "spire"
		}
		roleRef: {
			apiGroup: "rbac.authorization.k8s.io"
			kind:     "ClusterRole"
			name:     "system:openshift:scc:spire"
		}
		subjects: [{
			kind:      "ServiceAccount"
			name:      "agent"
			namespace: "spire"
		}]
	},
	// RoleBindings for Grey Matter services so they can access their agent.sock
	rbacv1.#RoleBinding & {// controlensemble
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "RoleBinding"
		metadata: {
			name:      "system:openshift:scc:spire:controlensemble"
			namespace: mesh.spec.install_namespace
		}
		roleRef: {
			apiGroup: "rbac.authorization.k8s.io"
			kind:     "ClusterRole"
			name:     "system:openshift:scc:spire"
		}
		subjects: [{
			kind:      "ServiceAccount"
			name:      "controlensemble"
			namespace: mesh.spec.install_namespace
		}]
	},
	rbacv1.#RoleBinding & {// default service account (used by all components except controlensemble)
		apiVersion: "rbac.authorization.k8s.io/v1"
		kind:       "RoleBinding"
		metadata: {
			name:      "system:openshift:scc:spire:default"
			namespace: mesh.spec.install_namespace
		}
		roleRef: {
			apiGroup: "rbac.authorization.k8s.io"
			kind:     "ClusterRole"
			name:     "system:openshift:scc:spire"
		}
		subjects: [{
			kind:      "ServiceAccount"
			name:      "default"
			namespace: mesh.spec.install_namespace
		}]
	},
]
