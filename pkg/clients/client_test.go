package clients

import (
	"testing"
)

func TestCLIVersion(t *testing.T) {
	v, err := cliVersion()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}
