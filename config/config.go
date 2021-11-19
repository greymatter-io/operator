package config

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"strings"
	"text/template"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

//go:embed *
var configFS embed.FS

type ManifestConfig struct {
	DockerImageURL               string
	DockerUsername               string
	DockerPassword               string
	DockerConfigBase64           string
	DisableWebhookCertGeneration bool
	ResourceLimitsCPU            string
	ResourceLimitsMemory         string
	ResourceRequestsCPU          string
	ResourceRequestsMemory       string
}

func LoadManifests(conf ManifestConfig) (string, error) {
	if (conf.DockerUsername == "" || conf.DockerPassword == "") && conf.DockerConfigBase64 == "" {
		return "", fmt.Errorf("missing docker credentials (either username and password or base64 dockercfgjson")
	}
	if conf.DockerImageURL == "" {
		conf.DockerImageURL = "docker.greymatter.io/development/gm-operator:0.0.1"
	}
	if conf.ResourceLimitsCPU == "" {
		conf.ResourceLimitsCPU = "200m"
	}
	if conf.ResourceLimitsMemory == "" {
		conf.ResourceLimitsMemory = "100Mi"
	}
	if conf.ResourceRequestsCPU == "" {
		conf.ResourceRequestsCPU = "100m"
	}
	if conf.ResourceRequestsMemory == "" {
		conf.ResourceRequestsMemory = "20Mi"
	}

	tmplString, err := loadTemplateString()
	if err != nil {
		return "", fmt.Errorf("failed to load template string: %w", err)
	}

	tmpl, err := template.New("manifests").Parse(tmplString)
	if err != nil {
		return "", fmt.Errorf("failed to parse template string: %w", err)
	}

	if conf.DockerConfigBase64 == "" {
		conf.DockerConfigBase64 = genDockerConfigBase64(conf.DockerUsername, conf.DockerPassword)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, conf); err != nil {
		return "", fmt.Errorf("failed to apply values to template: %w", err)
	}

	return buf.String(), nil
}

func genDockerConfigBase64(user, password string) string {
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", user, password)))
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`{
		"auths":{
			"docker.greymatter.io":{
				"username":"%s",
				"email":"%s",
				"password":"%s",
				"auth":"%s"
			}
		}
	}`, user, user, password, auth)))
}

func loadTemplateString() (string, error) {
	kfs, err := mkKyamlFileSys(configFS)
	if err != nil {
		return "", fmt.Errorf("failed to populate in-memory file system: %w", err)
	}

	opts := krusty.MakeDefaultOptions()
	opts.DoLegacyResourceSort = true

	k := krusty.MakeKustomizer(opts)
	res, err := k.Run(kfs, "context/kubernetes-options")
	if err != nil {
		return "", fmt.Errorf("failed to perform kustomization: %w", err)
	}

	yml, err := res.AsYaml()
	if err != nil {
		return "", fmt.Errorf("failed to parse as yaml")
	}

	return string(yml), nil
}

func mkKyamlFileSys(efs embed.FS) (filesys.FileSystem, error) {
	kfs := filesys.MakeFsInMemory()
	loadFunc := mkFileLoader(efs, kfs)
	if err := fs.WalkDir(efs, ".", loadFunc); err != nil {
		return kfs, err
	}
	return kfs, nil
}

func mkFileLoader(efs embed.FS, kfs filesys.FileSystem) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}
		data, err := efs.ReadFile(path)
		if err != nil {
			return err
		}
		return kfs.WriteFile(path, data)
	}
}
