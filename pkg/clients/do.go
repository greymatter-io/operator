package clients

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

func Do(httpClient *http.Client, action, url string, data []byte) ([]byte, error) {
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

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("client.Do: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
