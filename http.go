package wazobiamail

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// APIResponse is the standard WazobiaMail success envelope.
type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type requestOptions struct {
	idempotencyKey string
	query          url.Values
	headers        http.Header
}

func (c *Client) buildURL(path string, query url.Values) string {
	endpoint := c.baseURL + "/" + strings.TrimLeft(path, "/")
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	return endpoint
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader, opts requestOptions) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, opts.query), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	for name, values := range c.headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	if opts.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", opts.idempotencyKey)
	}
	for name, values := range opts.headers {
		for _, value := range values {
			req.Header.Set(name, value)
		}
	}
	return req, nil
}

func (c *Client) do(ctx context.Context, method, path string, payload any, opts requestOptions) ([]byte, http.Header, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, err
		}
		body = bytes.NewReader(encoded)
	}

	req, err := c.newRequest(ctx, method, path, body, opts)
	if err != nil {
		return nil, nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.Header, parseAPIError(resp.StatusCode, responseBody, resp.Header)
	}
	return responseBody, resp.Header, nil
}

func doJSON[T any](ctx context.Context, c *Client, method, path string, payload any, opts requestOptions) (*APIResponse[T], error) {
	body, _, err := c.do(ctx, method, path, payload, opts)
	if err != nil {
		return nil, err
	}
	var envelope APIResponse[T]
	if len(body) == 0 {
		return &envelope, nil
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

func (c *Client) doBytes(ctx context.Context, path string, opts requestOptions) ([]byte, http.Header, error) {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil, opts)
	if err != nil {
		return nil, nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.Header, parseAPIError(resp.StatusCode, body, resp.Header)
	}
	return body, resp.Header, nil
}
