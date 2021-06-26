package meshobjects

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/dougfort/traversal"
)

func raw(template string, values ...interface{}) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(template, values...))
}

func traverse(tr *traversal.Traversal) string {
	var buf bytes.Buffer
	tr.End(&buf)
	return buf.String()
}

func parseChecksum(s string) string {
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}
	return s
}
