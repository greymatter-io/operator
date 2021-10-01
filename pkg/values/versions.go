package values

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed versions/*.yaml
var filesystem embed.FS

// TODO: Allow the user to specify a directory for mounting new install files.
// This of course won't use embed.FS since those are embedded at compile time.
// User-provided values files must be validated on startup.
func LoadYAMLVersions() (map[string][]byte, error) {
	files, err := filesystem.ReadDir("versions")
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded template files: %w", err)
	}

	versions := make(map[string][]byte)
FILE_LOOP:
	for _, file := range files {
		fileName := file.Name()
		if !strings.HasSuffix(fileName, ".yaml") {
			logger.Error(fmt.Errorf("detected template file with invalid extension (expected .yaml)"), "skipping", "filename", fileName)
			continue FILE_LOOP
		}
		name := strings.Replace(fileName, ".yaml", "", 1)
		data, _ := filesystem.ReadFile(fmt.Sprintf("versions/%s", fileName))
		versions[name] = data
	}

	return versions, nil
}
