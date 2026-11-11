package wazobiamail

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
)

const (
	// WebhookSignatureHeader is the header that carries the hex HMAC SHA-256 signature.
	WebhookSignatureHeader = "X-Wazobia-Signature"
	// WebhookSecretEnv is the conventional environment variable for the signing secret.
	WebhookSecretEnv = "WAZOBIA_WEBHOOK_SECRET"
)

// ErrInvalidWebhookSignature is returned when a webhook signature does not match.
var ErrInvalidWebhookSignature = errors.New("wazobiamail: invalid webhook signature")

// WebhooksService verifies incoming webhook signatures without a network call.
type WebhooksService struct{}

// VerifySignature checks a hex HMAC SHA-256 signature. A "sha256=" prefix is accepted.
func (s *WebhooksService) VerifySignature(payload []byte, signature, secret string) bool {
	signature = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(signature), "sha256="))
	if signature == "" || secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(strings.ToLower(signature)))
}

// ConstructEvent verifies the signature and decodes the JSON payload into out.
func (s *WebhooksService) ConstructEvent(payload []byte, signature, secret string, out any) error {
	if !s.VerifySignature(payload, signature, secret) {
		return ErrInvalidWebhookSignature
	}
	return json.Unmarshal(payload, out)
}

// Sign returns the hex HMAC SHA-256 signature for a payload. Useful in tests.
func (s *WebhooksService) Sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
