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

//go:embed *
var configFS embed.FS

var KubernetesCommand = &cli.Command{
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "image",
			Usage:   "Which container image to use in the operator deployment.",
			Aliases: []string{"i"},
			Value:   "docker.greymatter.io/development/gm-operator:0.0.1",
		},
		&cli.StringFlag{
			Name:    "username",
			Usage:   "The username for accessing the Grey Matter container image repository",
			Aliases: []string{"u"},
		},
		&cli.StringFlag{
			Name:    "password",
			Usage:   "The password for accessing the Grey Matter container image repository",
			Aliases: []string{"p"},
		},
		&cli.BoolFlag{
			Name: "disable-internal-ca",
			Usage: "\n" + strings.Join([]string{
				"Disables the operator's internal certificate authority server. Note that the following must be manually configured if this flag is set:",
				"1. The Secret 'gm-controller-manager-service-cert' must have a signed 'tls.crt' and 'tls.key' with the SAN of 'gm-webhook-service.gm-operator.svc'.",
				"2. All webhooks defined in ValidatingWebhookConfiguration 'gm-validating-webhook-configuration' must have the signing CA cert in its .clientConfig.caBundle value.",
				"3. All webhooks defined in MutatingWebhookConfiguration 'gm-mutating-webhook-configuration' must have the signing CA cert in its .clientConfig.caBundle value.",
			}, "\n"),
			Value: false,
		},
	},
	Action: func(c *cli.Context) error {
		dockerUsername := c.String("username")
		dockerPassword := c.String("password")
		if dockerUsername == "" {
			dockerUsername = os.Getenv("GREYMATTER_DOCKER_USERNAME")
		}
		if dockerPassword == "" {
			dockerPassword = os.Getenv("GREYMATTER_DOCKER_PASSWORD")
		}
		return loadManifests("context/kubernetes-options", manifestConfig{
			DockerImageURL:               c.String("image"),
			DockerUsername:               dockerUsername,
			DockerPassword:               dockerPassword,
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
	if conf.DockerUsername == "" || conf.DockerPassword == "" {
		return fmt.Errorf("missing docker credentials")
	}
	conf.DockerConfigBase64 = genDockerConfigBase64(conf.DockerUsername, conf.DockerPassword)
	if conf.DockerImageURL == "" {
		conf.DockerImageURL = "docker.greymatter.io/development/gm-operator:0.0.1"
	}

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
