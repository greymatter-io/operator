package clients

import (
	"testing"
)

func TestVersion(t *testing.T) {
	v, err := version()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(v)
}

func TestHelp(t *testing.T) {
	v := help()
	t.Log(v)
}
