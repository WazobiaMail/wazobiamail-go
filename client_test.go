package wazobiamail

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	client, err := NewClient("wz_test_key", WithBaseURL(server.URL), WithTimeout(5_000_000_000))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return client
}

func writeJSON(w http.ResponseWriter, status int, value string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(value))
}

func TestNewClientRequiresCredentials(t *testing.T) {
	t.Setenv("WAZOBIA_API_KEY", "")
	t.Setenv("WAZOBIAMAIL_API_KEY", "")
	if _, err := NewClient(""); err == nil {
		t.Fatal("expected an error without credentials")
	}
	client, err := NewClient("", WithAllowUnauthenticated(true))
	if err != nil || client == nil {
		t.Fatalf("allow unauthenticated: %v", err)
	}
}

func TestClientAuthAndUserAgent(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-API-Key"); got != "wz_test_key" {
			t.Errorf("X-API-Key = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got != DefaultUserAgent {
			t.Errorf("User-Agent = %q", got)
		}
		writeJSON(w, http.StatusOK, `{"success":true,"data":{"ok":true}}`)
	})
	var out struct {
		OK bool `json:"ok"`
	}
	if _, err := doJSON[struct {
		OK bool `json:"ok"`
	}](context.Background(), client, http.MethodGet, "/ping", nil, requestOptions{}); err != nil {
		t.Fatalf("doJSON: %v", err)
	}
	_ = out
}

func TestAPIError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "7")
		writeJSON(w, http.StatusUnprocessableEntity, `{"message":"Validation failed","code":"validation_error","errors":{"to":"required"}}`)
	})
	_, err := client.Messages.Send(context.Background(), SendMessageRequest{To: []string{"a@example.com"}, Subject: "Hi", Text: "Hello"})
	if err == nil {
		t.Fatal("expected an error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type %T is not *APIError", err)
	}
	if !apiErr.IsValidation() || apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unexpected API error: %+v", apiErr)
	}
	if apiErr.Message != "Validation failed" || apiErr.Code != "validation_error" || apiErr.Errors["to"] != "required" {
		t.Fatalf("error fields missing: %+v", apiErr)
	}
	if apiErr.RetryAfter.Seconds() != 7 {
		t.Fatalf("RetryAfter = %v", apiErr.RetryAfter)
	}
}

func TestWithAccessToken(t *testing.T) {
	client, err := NewClient("", WithAccessToken("bearer-token"))
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer bearer-token" {
			t.Errorf("Authorization = %q", got)
		}
		writeJSON(w, http.StatusOK, `{"success":true,"data":{}}`)
	}))
	defer server.Close()
	client.baseURL = server.URL
	if _, err := doJSON[map[string]any](context.Background(), client, http.MethodGet, "/ping", nil, requestOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestQueryEncoding(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("limit"); got != "10" {
			t.Errorf("limit = %q", got)
		}
		if got := r.URL.Query().Get("status"); got != "delivered" {
			t.Errorf("status = %q", got)
		}
		writeJSON(w, http.StatusOK, `{"success":true,"data":{"messages":[],"summary":{}}}`)
	})
	_, err := client.Messages.ListTracked(context.Background(), TrackingListParams{Limit: 10, Status: "delivered"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidJSONError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, `{not-json`)
	})
	_, err := client.Messages.SpamCheck(context.Background(), SpamCheckRequest{Subject: "Hi"})
	if err == nil {
		t.Fatal("expected JSON error")
	}
	var syntax *json.SyntaxError
	if !errors.As(err, &syntax) {
		t.Fatalf("expected json.SyntaxError, got %T", err)
	}
}

func TestStringMasksKey(t *testing.T) {
	client, _ := NewClient("wz_live_123456789")
	if got := client.String(); got == "" || contains(got, "123456789") {
		t.Fatalf("client string leaks key: %q", got)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
