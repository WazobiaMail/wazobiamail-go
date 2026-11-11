package wazobiamail

import (
	"errors"
	"testing"
)

func TestWebhookVerifyAndConstruct(t *testing.T) {
	client, err := NewClient("wz_test", WithAllowUnauthenticated(true))
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"type":"email.delivered","data":{"trackingId":"trk_1"}}`)
	secret := "whsec_test"
	signature := client.Webhooks.Sign(payload, secret)

	if !client.Webhooks.VerifySignature(payload, signature, secret) {
		t.Fatal("valid signature rejected")
	}
	if !client.Webhooks.VerifySignature(payload, "sha256="+signature, secret) {
		t.Fatal("sha256-prefixed signature rejected")
	}
	if client.Webhooks.VerifySignature(payload, signature, "wrong") {
		t.Fatal("wrong secret accepted")
	}

	var event struct {
		Type string `json:"type"`
		Data struct {
			TrackingID string `json:"trackingId"`
		} `json:"data"`
	}
	if err := client.Webhooks.ConstructEvent(payload, signature, secret, &event); err != nil {
		t.Fatal(err)
	}
	if event.Type != "email.delivered" || event.Data.TrackingID != "trk_1" {
		t.Fatalf("event = %+v", event)
	}

	err = client.Webhooks.ConstructEvent(payload, "bad", secret, &event)
	if !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Fatalf("error = %v", err)
	}
}
