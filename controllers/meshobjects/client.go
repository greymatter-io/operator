package meshobjects

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	addr       string
}

func NewClient(addr string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: time.Second},
		addr:       fmt.Sprintf("%s/v1.0", addr),
	}
}

func (c *Client) do(action, url string, data []byte) ([]byte, error) {
	var req *http.Request
	var err error

	switch action {
	case http.MethodGet:
		req, err = http.NewRequest(http.MethodGet, url, nil)
	default:
		req, err = http.NewRequest(action, url, bytes.NewBuffer(data))
	}
	if err != nil {
		return nil, fmt.Errorf("http.NewRequest: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apiClient.Do(req): %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to parse response body from POST %s: %s", url, err.Error())
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("%s %s responded with status %s", action, url, resp.Status)
		log.Print(string(body))
		return body, err
	}

	return body, nil
}
