package fabric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestFabric(t *testing.T) {
	f, err := New("default-zone", 10808)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("f.Edge returns edge objects", func(t *testing.T) {
		edge := f.Edge()
		fmt.Println(indent(edge.Proxy))
		fmt.Println(indent(edge.Domain))
		fmt.Println(indent(edge.Listener))
		fmt.Println(indent(edge.Cluster))
	})

	t.Run("f.Service returns service objects", func(t *testing.T) {
		service := f.Service("blah", []int32{5555, 8080})

		fmt.Println(indent(service.Proxy))
		fmt.Println(indent(service.Domain))
		fmt.Println(indent(service.Listener))
		fmt.Println(indent(service.Cluster))
		fmt.Println(indent(service.Route))
		for _, local := range service.Locals {
			fmt.Println(indent(local.Cluster))
			fmt.Println(indent(local.Route))
		}
	})
}

func indent(raw json.RawMessage) string {
	b := new(bytes.Buffer)
	json.Indent(b, raw, "", "\t")
	return b.String()
}
