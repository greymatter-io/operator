package config

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestKubernetesCommand(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{KubernetesCommand}
	if err := app.Run([]string{"", "",
		"--image", "my-docker-image-url",
		"--username", "my-docker-user",
		"--password", "my-docker-password",
		"--disable-internal-ca",
	}); err != nil {
		t.Error(err)
	}
}

func TestKubernetesCommandDockerAuthEnvVars(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{KubernetesCommand}
	os.Setenv("GREYMATTER_DOCKER_USERNAME", "my-docker-user")
	os.Setenv("GREYMATTER_DOCKER_PASSWORD", "my-docker-password")
	if err := app.Run([]string{"", ""}); err != nil {
		t.Error(err)
	}
}

func TestKubernetesCommandHelp(t *testing.T) {
	app := cli.NewApp()
	KubernetesCommand.Name = "cmd"
	app.Commands = []*cli.Command{KubernetesCommand}
	if err := app.Run([]string{"", "cmd", "-h"}); err != nil {
		t.Error(err)
	}
}

func TestLoadManifests(t *testing.T) {
	conf := manifestConfig{
		DockerImageURL:               "my-docker-image-url",
		DockerUsername:               "my-docker-user",
		DockerPassword:               "my-docker-password",
		DisableWebhookCertGeneration: true,
	}
	if err := loadManifests("context/kubernetes-options", conf); err != nil {
		t.Fatal(err)
	}
}

func TestLoadTemplateString(t *testing.T) {
	tmplStr, err := loadTemplateString("context/kubernetes-options")
	if err != nil {
		t.Fatal(err)
	}

	for _, placeholder := range []string{
		"DockerImageURL",
		"DisableWebhookCertGeneration",
		"DockerConfigBase64",
	} {
		if !strings.Contains(tmplStr, fmt.Sprintf("{{ .%s }}", placeholder)) {
			t.Errorf("Did not find placeholder for %s", placeholder)
		}
	}
}

func TestMkKyamlFileSys(t *testing.T) {
	kfs, err := mkKyamlFileSys(configFS)
	if err != nil {
		t.Fatal(err)
	}

	// Walk the embed.FS and ensure all YAML files exist in the kyaml file system
	err = fs.WalkDir(configFS, ".", func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".yaml") {
			if !kfs.Exists(path) {
				return fmt.Errorf("kfs does not have file %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	// Walk the kyaml file system and ensure all files exist in embed.FS
	err = kfs.Walk("/", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if _, err := configFS.Open(strings.TrimPrefix(path, "/")); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
