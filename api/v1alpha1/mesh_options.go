package v1alpha1

type UserToken struct {
	Label  string              `json:"label"`
	Values map[string][]string `json:"values"`
}

// ImageSecret can be defined on a per-image basis. The secret name,
// as well as the secret host namespace are required due to the operator
// being segmented in its own namespace.
type ImageSecret struct {
	Name      string `json:"secret_name"`
	Namespace string `json:"secret_namespace"`
}
