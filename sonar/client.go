package sonar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"open-sonar/internal/models"
)

// Client is a Go client for the Open Sonar API
type Client struct {
	BaseURL    string
	AuthToken  string
	HTTPClient *http.Client
}

// NewClient creates a new Open Sonar API client
func NewClient(baseURL string, authToken string) *Client {
	return &Client{
		BaseURL:    baseURL,
		AuthToken:  authToken,
		HTTPClient: &http.Client{},
	}
}

// Chat sends a simple chat request
func (c *Client) Chat(query string, needSearch bool) (map[string]interface{}, error) {
	req := map[string]interface{}{
		"query":      query,
		"needSearch": needSearch,
	}

	var resp map[string]interface{}
	err := c.sendRequest("POST", "/chat", req, &resp)
	return resp, err
}

// ChatCompletions sends a chat completions request (OpenAI compatible)
func (c *Client) ChatCompletions(request models.ChatCompletionRequest) (*models.ChatCompletionResponse, error) {
	var resp models.ChatCompletionResponse
	err := c.sendRequest("POST", "/chat/completions", request, &resp)
	return &resp, err
}

// sendRequest sends an HTTP request to the API
func (c *Client) sendRequest(method, path string, body interface{}, result interface{}) error {
	// Marshal the request body
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	// Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Decode response
	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
	}

	return nil
}
