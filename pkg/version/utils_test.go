package version

import (
	"fmt"
	"strings"
	"testing"
)

func TestCueFromStrings(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []string
	}{
		{
			name: "one",
			in:   []string{"a: 1"},
		},
		{
			name: "two",
			in:   []string{"a: 1", "b: 2"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			value, err := CueFromStrings(tc.in...)
			if err != nil {
				t.Fatal(err)
			}
			out := fmt.Sprintf("%v", value)
			for _, expr := range tc.in {
				if !strings.Contains(out, expr) {
					t.Errorf("expected to find %s in cue.Value", expr)
				}
			}
		})
	}
}
