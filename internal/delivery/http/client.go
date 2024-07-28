// client.go
package request

import (
	"net/http"
)

// HTTPClient interface defines the methods required for making HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// JSONRequester interface defines methods for making JSON requests
type JSONRequester interface {
	Post(url string, data interface{}) (*http.Response, error)
}

// CustomHTTPClient implements the HTTPClient interface
type CustomHTTPClient struct {
	Client *http.Client
}

// NewCustomHTTPClient creates a new CustomHTTPClient
func NewCustomHTTPClient() *CustomHTTPClient {
	return &CustomHTTPClient{
		Client: &http.Client{},
	}
}

// Do implements the Do method of the HTTPClient interface
func (c *CustomHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}
