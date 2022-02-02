package config

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

var testConf = manifestConfig{
	DockerImageURL:               "my-docker-image-url",
	DockerUsername:               "my-docker-user",
	DockerPassword:               "my-docker-password",
	DisableWebhookCertGeneration: true,
	ImagePullSecretsList:         []string{"secret1", "secret2"},
}

func TestKubernetesCommand(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{&kubernetesCommand}
	if err := app.Run([]string{"", "",
		"--image", testConf.DockerImageURL,
		"--registry-username", testConf.DockerUsername,
		"--registry-password", testConf.DockerPassword,
		"--pull-secrets", strings.Join(testConf.ImagePullSecretsList, ","),
		"--disable-internal-ca",
	}); err != nil {
		t.Error(err)
	}
}

func TestKubernetesCommandDockerAuthEnvVars(t *testing.T) {
	app := cli.NewApp()
	app.Commands = []*cli.Command{&kubernetesCommand}
	os.Setenv("GREYMATTER_REGISTRY_USERNAME", testConf.DockerUsername)
	os.Setenv("GREYMATTER_REGISTRY_PASSWORD", testConf.DockerPassword)
	if err := app.Run([]string{"", ""}); err != nil {
		t.Error(err)
	}
}

func TestKubernetesCommandHelp(t *testing.T) {
	app := cli.NewApp()
	kubernetesCommand.Name = "cmd"
	app.Commands = []*cli.Command{&kubernetesCommand}
	if err := app.Run([]string{"", "cmd", "-h"}); err != nil {
		t.Error(err)
	}
}

func TestLoadManifests(t *testing.T) {
	if err := loadManifests("context/kubernetes-options", testConf); err != nil {
		t.Fatal(err)
	}
}

func TestLoadTemplateString(t *testing.T) {
	tc := testConf
	// Manually set the base64 Docker config since it's generated outside this function
	tc.DockerConfigBase64 = genDockerConfigBase64(tc.DockerUsername, tc.DockerPassword)

	tms, err := loadTemplatedManifests("context/kubernetes-options", tc)
	if err != nil {
		t.Fatal(err)
	}

	expectedValues := append([]string{
		tc.DockerImageURL,
		tc.DockerConfigBase64,
	}, tc.ImagePullSecretsList...)

	for _, value := range expectedValues {
		if !strings.Contains(tms, value) {
			t.Errorf("Did not find value %s", value)
		}
	}
}

func TestMkKyamlFileSys(t *testing.T) {
	kfs, err := mkKyamlFileSys(configFS, testConf)
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
