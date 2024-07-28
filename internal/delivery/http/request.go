// requester.go
package request

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// CustomJSONRequester implements the JSONRequester interface
type CustomJSONRequester struct {
	Client HTTPClient
}

// NewCustomJSONRequester creates a new CustomJSONRequester
func NewCustomJSONRequester(client HTTPClient) *CustomJSONRequester {
	return &CustomJSONRequester{
		Client: client,
	}
}

// Post implements the Post method of the JSONRequester interface
func (r *CustomJSONRequester) Post(url string, data interface{}) (*http.Response, error) {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Create a new HTTP request with POST method and JSON data in the body
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Set the Content-Type header to application/json
	req.Header.Set("Content-Type", "application/json")

	// Make the request
	return r.Client.Do(req)
}
