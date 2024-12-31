package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/Easy-Infra-Ltd/easy-test/internal/assert"
)

type ClientConfig struct {
	Url         string `json:"url"`
	ContentType string `json:"contentType"`
}

type ClientParams struct {
	url         string
	contentType string
	body        io.Reader
}

func NewClientParams(url string, contentType string, body io.Reader) *ClientParams {
	assert.NotNil(url, "Client params url can not be nil")
	assert.NotNil(contentType, "Client params contentType cannot be nil")

	return &ClientParams{
		url:         url,
		contentType: contentType,
		body:        body,
	}
}

type Client struct {
	logger *slog.Logger
	config *ClientParams
}

func NewClient(config *ClientParams) *Client {
	logger := slog.Default().With("area", "HTTPClient")
	return &Client{
		logger: logger,
		config: config,
	}
}

func (c *Client) Post() (*http.Response, error) {
	c.logger.Info(fmt.Sprintf("Sending POST request to url %s with contentType %s", c.config.url, c.config.contentType))
	return http.Post(c.config.url, c.config.contentType, c.config.body)
}

func (c *Client) Get() (*http.Response, error) {
	c.logger.Info(fmt.Sprintf("Sending GET request to url %s with contentType %s", c.config.url, c.config.contentType))
	return http.Get(c.config.url)
}
