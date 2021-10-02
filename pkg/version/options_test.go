package version

import (
	"testing"

	"cuelang.org/go/cue/errors"
)

func TestOptions(t *testing.T) {
	versions, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	for name, version := range versions {
		t.Run(name, func(t *testing.T) {
			t.Run("SPIRE", func(t *testing.T) {
				vCopy := version.Copy()
				vCopy.Apply(SPIRE)
				if err := vCopy.cv.Err(); err != nil {
					for _, err := range errors.Errors(err) {
						t.Error(err)
					}
				}
				// TODO: Validate struct
			})

			t.Run("InternalRedis", func(t *testing.T) {
				vCopy := version.Copy()
				vCopy.Apply(InternalRedis("namespace"))
				if err := vCopy.cv.Err(); err != nil {
					for _, err := range errors.Errors(err) {
						t.Error(err)
					}
				}
				// TODO: Validate struct
			})

			t.Run("ExternalRedis", func(t *testing.T) {
				vCopy := version.Copy()
				vCopy.Apply(ExternalRedis(&ExternalRedisConfig{URL: "redis://:pass@extserver:6379/2"}))
				if err := vCopy.cv.Err(); err != nil {
					for _, err := range errors.Errors(err) {
						t.Error(err)
					}
				}
				// TODO: Validate struct
			})
		})
	}
}
