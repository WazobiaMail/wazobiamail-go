package wazobiamail

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
)

func (c *Client) doMultipart(ctx context.Context, path string, write func(*multipart.Writer) error, opts requestOptions) ([]byte, http.Header, error) {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)
	if err := write(writer); err != nil {
		return nil, nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, nil, err
	}

	req, err := c.newRequest(ctx, http.MethodPost, path, &buffer, opts)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	var responseBody bytes.Buffer
	if _, err := responseBody.ReadFrom(resp.Body); err != nil {
		return nil, resp.Header, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.Header, parseAPIError(resp.StatusCode, responseBody.Bytes(), resp.Header)
	}
	return responseBody.Bytes(), resp.Header, nil
}

func decodeBody[T any](body []byte) (*APIResponse[T], error) {
	var envelope APIResponse[T]
	if len(body) == 0 {
		return &envelope, nil
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
