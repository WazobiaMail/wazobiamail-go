package wazobiamail

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// APIError is returned for non-2xx API responses.
type APIError struct {
	StatusCode int
	Message    string
	Code       string
	Errors     map[string]any
	Body       []byte
	RetryAfter time.Duration
}

// Error implements error.
func (e *APIError) Error() string {
	if e == nil {
		return "wazobiamail: <nil>"
	}
	message := strings.TrimSpace(e.Message)
	if message == "" {
		message = "API request failed"
	}
	return fmt.Sprintf("wazobiamail: %s (HTTP %d)", message, e.StatusCode)
}

// IsNotFound reports whether the API returned 404.
func (e *APIError) IsNotFound() bool { return e != nil && e.StatusCode == 404 }

// IsUnauthorized reports whether the API returned 401.
func (e *APIError) IsUnauthorized() bool { return e != nil && e.StatusCode == 401 }

// IsForbidden reports whether the API returned 403.
func (e *APIError) IsForbidden() bool { return e != nil && e.StatusCode == 403 }

// IsValidation reports whether the API returned 422 (or a 400 with field errors).
func (e *APIError) IsValidation() bool {
	return e != nil && (e.StatusCode == 422 || (e.StatusCode == 400 && len(e.Errors) > 0))
}

// IsRateLimited reports whether the API returned 429.
func (e *APIError) IsRateLimited() bool { return e != nil && e.StatusCode == 429 }

func parseAPIError(status int, body []byte, header map[string][]string) *APIError {
	apiErr := &APIError{StatusCode: status, Body: body}
	var payload struct {
		Message string         `json:"message"`
		Error   any            `json:"error"`
		Code    string         `json:"code"`
		Errors  map[string]any `json:"errors"`
	}
	if json.Unmarshal(body, &payload) == nil {
		apiErr.Message = strings.TrimSpace(payload.Message)
		if apiErr.Message == "" {
			switch value := payload.Error.(type) {
			case string:
				apiErr.Message = value
			case map[string]any:
				if detail, ok := value["message"].(string); ok {
					apiErr.Message = detail
				}
			}
		}
		apiErr.Code = payload.Code
		apiErr.Errors = payload.Errors
	}
	if apiErr.Message == "" {
		apiErr.Message = httpStatusText(status)
	}
	if values := header["Retry-After"]; len(values) > 0 {
		if seconds, err := strconv.Atoi(strings.TrimSpace(values[0])); err == nil && seconds > 0 {
			apiErr.RetryAfter = time.Duration(seconds) * time.Second
		}
	}
	return apiErr
}

func httpStatusText(status int) string {
	switch status {
	case 400:
		return "Bad request"
	case 401:
		return "Authentication failed"
	case 403:
		return "Permission denied"
	case 404:
		return "Resource not found"
	case 409:
		return "Resource conflict"
	case 422:
		return "Validation failed"
	case 429:
		return "Rate limit exceeded"
	default:
		if status >= 500 {
			return "Server error"
		}
		return fmt.Sprintf("HTTP %d", status)
	}
}
