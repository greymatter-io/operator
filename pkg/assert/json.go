package assert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func JSONHasSubstrings(obj json.RawMessage, subs ...string) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()
		if len(obj) == 0 {
			t.Fatal("json is empty")
		}
		for _, sub := range subs {
			if !bytes.Contains(obj, json.RawMessage(sub)) {
				t.Errorf("did not contain substring '%s'", sub)
			}
		}
		if t.Failed() {
			prettyPrintJSON(obj)
		}
	}
}

func prettyPrintJSON(raws ...json.RawMessage) {
	for _, raw := range raws {
		b := new(bytes.Buffer)
		json.Indent(b, raw, "", "\t")
		fmt.Println(b.String())
	}
}
