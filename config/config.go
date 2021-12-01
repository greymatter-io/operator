package config

import (
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
			Name:     "username",
			Usage:    "The username for accessing the Grey Matter container image repository.",
			Aliases:  []string{"u"},
			EnvVars:  []string{"GREYMATTER_DOCKER_USERNAME"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "password",
			Usage:    "The password for accessing the Grey Matter container image repository.",
			Aliases:  []string{"p"},
			EnvVars:  []string{"GREYMATTER_DOCKER_PASSWORD"},
			Required: true,
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
			DockerUsername:               c.String("username"),
			DockerPassword:               c.String("password"),
			DisableWebhookCertGeneration: c.Bool("disable-internal-ca"),
		})
	},
}

type manifestConfig struct {
	DockerImageURL               string
	DockerUsername               string
	DockerPassword               string
	DisableWebhookCertGeneration bool
	// Generated from DockerUsername and DockerPassword
	DockerConfigBase64 string
}

func loadManifests(dirPath string, conf manifestConfig) error {
	conf.DockerConfigBase64 = genDockerConfigBase64(conf.DockerUsername, conf.DockerPassword)

	tmplString, err := loadTemplateString(dirPath)
	if err != nil {
		return fmt.Errorf("failed to load template string: %w", err)
	}

	tmpl, err := template.New("manifests").Parse(tmplString)
	if err != nil {
		return fmt.Errorf("failed to parse template string: %w", err)
	}

	return tmpl.Execute(os.Stdout, conf)
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

func loadTemplateString(dirPath string) (string, error) {
	kfs, err := mkKyamlFileSys(configFS)
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
func mkKyamlFileSys(efs embed.FS) (filesys.FileSystem, error) {
	kfs := filesys.MakeFsInMemory()
	loadFunc := mkFileLoader(efs, kfs)
	if err := fs.WalkDir(efs, ".", loadFunc); err != nil {
		return nil, err
	}
	return kfs, nil
}

// Receives an embed.FS (from the Go 1.16+ standard library) and an filesys.FileSystem (from kustomize)
// and returns a function that implements the fs.WalkDirFunc function signature for each embed.FS fs.DirEntry.
// The returned function reads YAML files from the embed.FS and writes it to the same path in the filesys.FileSystem.
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
