package meshobjects

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dougfort/traversal"
)

func (c *Client) Ping() error {
	url := fmt.Sprintf("%s/zone", c.addr)

	var err error
	for i := 0; i < 5; i++ {
		if _, err := c.do(http.MethodGet, url, nil); err != nil {
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}
	if err != nil {
		log.Printf("failed to ping Control API: %s", err.Error())
		return errors.New("ping")
	}

	return nil
}

func (c *Client) Make(kind, key string, object json.RawMessage) error {
	url := fmt.Sprintf("%s/%s", c.addr, kind)

	if _, err := c.do(http.MethodPost, url, object); err != nil {
		return err
	}

	log.Printf("Created %s '%s'", kind, key)

	return nil
}

func (c *Client) Change(kind, key string, changes map[string]json.RawMessage) error {
	url := fmt.Sprintf("%s/%s/%s", c.addr, kind, key)

	body, err := c.do(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	bodyMap, err := traversal.GetMapFromRawMessage(body)
	if err != nil {
		return fmt.Errorf("traversal.GetMapFromRawMessage: %w", err)
	}

	result, ok := bodyMap["result"]
	if !ok {
		return fmt.Errorf("GET %s did not have 'result', got %s", url, string(body))
	}

	object, err := traversal.GetMapFromRawMessage(result)
	if err != nil {
		return fmt.Errorf("traversal.GetMapFromRawMessage: %w", err)
	}

	for k, v := range changes {
		object[k] = v
	}

	updated, err := json.Marshal(object)
	if err != nil {
		return err
	}

	if _, err := c.do("PUT", url, updated); err != nil {
		return err
	}

	log.Printf("Updated %s '%s'", kind, key)

	return nil
}
