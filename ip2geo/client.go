package ip2geo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.ip2geoapi.com/ip"

type Client struct {
	APIKey  string
	Timeout time.Duration
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:  apiKey,
		Timeout: 60 * time.Second,
	}
}

func (c *Client) Lookup(ip string, format string, callback string) (interface{}, error) {
	params := url.Values{}

	if c.APIKey != "" {
		params.Add("key", c.APIKey)
	}

	if format != "" {
		params.Add("format", format)
	}

	if callback != "" {
		if format != "jsonp" {
			return nil, errors.New("callback can only be used when format is 'jsonp'")
		}
		params.Add("callback", callback)
	}

	endpoint := baseURL
	if ip != "" {
		endpoint = fmt.Sprintf("%s/%s", baseURL, ip)
	}

	reqURL := endpoint
	if len(params) > 0 {
		reqURL = fmt.Sprintf("%s?%s", endpoint, params.Encode())
	}

	client := &http.Client{Timeout: c.Timeout}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Default JSON handling
	if format == "" || format == "json" {
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, err
		}

		if success, ok := data["success"].(bool); ok && !success {
			if msg, ok := data["error"].(string); ok {
				return nil, errors.New(msg)
			}
			return nil, errors.New("unknown API error")
		}

		return data, nil
	}

	// Non-JSON formats return raw string
	return string(body), nil
}
