package v1alpha1

type RedisConfig struct {
	Url        string `json:"url"`
	SecretName string `json:"certificateSecretName"`
}
