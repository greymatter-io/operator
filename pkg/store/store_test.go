package store

import (
	"testing"

	"github.com/greymatter-io/operator/api/v1alpha1"
)

func TestCheckRedisConfigWorking(t *testing.T) {
	rc := v1alpha1.RedisConfig{
		Url:        "localhost:9999",
		SecretName: "",
	}

	ms := v1alpha1.MeshSpec{
		Foo:         "bar",
		RedisConfig: &rc,
	}
	err := checkRedisConfig(ms)
	if err != nil {
		t.Fail()
	}
}

func TestCheckRedisConfigNoRedisConfit(t *testing.T) {
	ms := v1alpha1.MeshSpec{
		Foo: "bar",
	}
	err := checkRedisConfig(ms)
	if err == nil {
		t.Fail()
	}
}

func TestCheckRedisConfigNoUrlSet(t *testing.T) {
	rc := v1alpha1.RedisConfig{
		Url:        "",
		SecretName: "",
	}

	ms := v1alpha1.MeshSpec{
		Foo:         "bar",
		RedisConfig: &rc,
	}
	err := checkRedisConfig(ms)
	if err != nil {
		t.Fail()
	}
	t.Logf("Url is set to: %s", rc.Url)
	if rc.Url != "redis://greymatteroperator:redis@localhost:6379/0" {
		t.Fail()
	}
}
