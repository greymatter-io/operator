package config

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

const OperatorImageURL = "docker.greymatter.io/development/gm-operator:latest"

//go:embed *
var configFS embed.FS

func MkKubernetesCommand(name, usage string) *cli.Command {
	command := kubernetesCommand
	command.Name = name
	command.Usage = usage
	return &command
}

var kubernetesCommand = cli.Command{
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "image",
			Usage:   "Which container image to use in the operator deployment.",
			Aliases: []string{"i"},
			Value:   OperatorImageURL,
		},
		&cli.StringFlag{
			Name:     "registry-username",
			Usage:    "The username for accessing the Grey Matter container image repository.",
			EnvVars:  []string{"GREYMATTER_REGISTRY_USERNAME"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "registry-password",
			Usage:    "The password for accessing the Grey Matter container image repository.",
			EnvVars:  []string{"GREYMATTER_REGISTRY_PASSWORD"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "pull-secrets",
			Usage:   "A command delimited list of known image pull secrets to use for fetching core services.",
			EnvVars: []string{"GREYMATTER_PULL_SECRETS_LIST"},
		},
		&cli.BoolFlag{
			Name: "disable-internal-ca",
			Usage: strings.Join([]string{
				"Disables the operator's internal certificate authority server. Note that the following must be manually configured if this flag is set:",
				"\t1. The Secret 'gm-webhook-cert' must have a signed 'tls.crt' and 'tls.key' with the SAN of 'gm-webhook.gm-operator.svc'.",
				"\t2. All webhooks defined in ValidatingWebhookConfiguration 'gm-validate-config' must have the signing CA cert in its .clientConfig.caBundle value.",
				"\t3. All webhooks defined in MutatingWebhookConfiguration 'gm-mutate-config' must have the signing CA cert in its .clientConfig.caBundle value.\n\t",
			}, "\n"),
			Value: false,
		},
	},
	Action: func(c *cli.Context) error {
		return loadManifests("context/kubernetes-options", manifestConfig{
			DockerImageURL:               c.String("image"),
			DockerUsername:               c.String("registry-username"),
			DockerPassword:               c.String("registry-password"),
			DisableWebhookCertGeneration: c.Bool("disable-internal-ca"),
			SecretsList:                  strings.Split(c.String("pull-secrets"), ","),
		})
	},
}

// manifestConfig contains options read from CLI flags
// to be used in the applied kustomize template patches.
type manifestConfig struct {
	DockerImageURL               string
	DockerUsername               string
	DockerPassword               string
	DisableWebhookCertGeneration bool

	// Generated from DockerUsername and DockerPassword
	DockerConfigBase64 string
	// SecretsList contains a comma-delimited slice of known imagePullSecret names
	SecretsList []string
}

func loadManifests(dirPath string, conf manifestConfig) error {
	conf.DockerConfigBase64 = genDockerConfigBase64(conf.DockerUsername, conf.DockerPassword)

	result, err := loadTemplatedManifests(dirPath, conf)
	if err != nil {
		return fmt.Errorf("failed to load templated manifests: %w", err)
	}

	_, err = os.Stdout.WriteString(result)
	return err
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

func loadTemplatedManifests(dirPath string, conf manifestConfig) (string, error) {
	kfs, err := mkKyamlFileSys(configFS, conf)
	if err != nil {
		return "", fmt.Errorf("failed to populate in-memory file system: %w", err)
	}

	opts := krusty.MakeDefaultOptions()
	opts.DoLegacyResourceSort = true

	k := krusty.MakeKustomizer(opts)
	res, err := k.Run(kfs, dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to perform kustomization: %w", err)
	}

	yml, err := res.AsYaml()
	if err != nil {
		return "", fmt.Errorf("failed to parse as yaml")
	}

	return string(yml), nil
}

// Loads a filesys.FileSystem with data from an embed.FS suitable for building a kustomize resource map.
func mkKyamlFileSys(efs embed.FS, conf manifestConfig) (filesys.FileSystem, error) {
	kfs := filesys.MakeFsInMemory()
	loadFunc := mkFileLoader(efs, kfs, conf)
	if err := fs.WalkDir(efs, ".", loadFunc); err != nil {
		return nil, err
	}
	return kfs, nil
}

// Receives an embed.FS (from the Go 1.16+ standard library) and an filesys.FileSystem (from kustomize)
// and returns a function that implements the fs.WalkDirFunc function signature for each embed.FS fs.DirEntry.
// The returned function reads YAML files from the embed.FS and writes it to the same path in the filesys.FileSystem.
func mkFileLoader(efs embed.FS, kfs filesys.FileSystem, conf manifestConfig) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		// Load the YAML file from the embed.FS.
		data, err := efs.ReadFile(path)
		if err != nil {
			return err
		}

		// If the YAML file is in our template directory, execute it with our manifestConfig
		// and overwrite the template in memory with the result.
		if strings.HasPrefix(path, "context/kubernetes-options/") &&
			!strings.HasSuffix(path, "kustomization.yaml") {

			tmpl, err := template.New(path).Parse(string(data))
			if err != nil {
				return fmt.Errorf("failed to parse: %w", err)
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, conf); err != nil {
				return fmt.Errorf("failed to execute: %w", err)
			}

			data = buf.Bytes()
		}

		return kfs.WriteFile(path, data)
	}
}
