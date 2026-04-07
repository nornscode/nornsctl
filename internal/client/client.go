package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func New(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL:    baseURL + "/api/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) do(method, path string, body any) ([]byte, error) {
	url := c.BaseURL + path

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(data, &apiErr) == nil && apiErr.Error != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, apiErr.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(data))
	}

	return data, nil
}

func (c *Client) Get(path string) ([]byte, error) {
	return c.do(http.MethodGet, path, nil)
}

func (c *Client) Post(path string, body any) ([]byte, error) {
	return c.do(http.MethodPost, path, body)
}

func (c *Client) Put(path string, body any) ([]byte, error) {
	return c.do(http.MethodPut, path, body)
}

func (c *Client) Delete(path string) ([]byte, error) {
	return c.do(http.MethodDelete, path, nil)
}
