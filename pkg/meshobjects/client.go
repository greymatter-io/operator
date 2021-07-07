package meshobjects

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

type Client struct {
	httpClient *http.Client
	addr       string
	logger     logr.Logger
}

func NewClient(addr string, logger logr.Logger) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: time.Second * 3},
		addr:       fmt.Sprintf("%s/v1.0", addr),
		logger:     logger,
	}
}
