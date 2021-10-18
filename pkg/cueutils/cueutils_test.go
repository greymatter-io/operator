package cueutils

import (
	"fmt"
	"strings"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestFromStrings(t *testing.T) {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

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
			value := FromStrings(tc.in...)
			if err := value.Err(); err != nil {
				LogError(ctrl.Log, err)
				t.FailNow()
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
