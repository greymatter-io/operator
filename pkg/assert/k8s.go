package assert

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func ContainerHasEnvValues(c corev1.Container, envs map[string]string) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		actual := make(map[string]string)
		for _, e := range c.Env {
			actual[e.Name] = e.Value
		}

		for k, v := range envs {
			if av, ok := actual[k]; !ok {
				t.Errorf("missing key '%s'", k)
			} else if av != v {
				t.Errorf("expected '%s' to be '%s' but got '%s'", k, v, av)
			}
		}
	}
}
