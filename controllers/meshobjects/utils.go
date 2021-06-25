package meshobjects

import (
	"encoding/json"
	"fmt"
)

func raw(template string, values ...interface{}) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(template, values...))
}
