package wazobiamail

import (
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"
)

func TestSendBuildsNestedPayload(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/sending/messages" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Idempotency-Key"); got != "order-123" {
			t.Errorf("Idempotency-Key = %q", got)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload["to"].([]any)[0] != "ada@example.com" {
			t.Errorf("to = %#v", payload["to"])
		}
		content := payload["content"].(map[string]any)
		if content["subject"] != "Your receipt" || content["htmlBody"] != "<p>Thanks</p>" || content["replyTo"] != "support@example.com" {
			t.Errorf("content = %#v", content)
		}
		if payload["envelopeSender"] != "billing@example.com" {
			t.Errorf("envelopeSender = %#v", payload["envelopeSender"])
		}
		writeJSON(w, http.StatusOK, `{"success":true,"message":"queued","data":{"result":{"status":"queued","trackingId":"trk_123","recipients":2}}}`)
	})
	resp, err := client.Messages.Send(context.Background(), SendMessageRequest{
		To:             []string{"ada@example.com"},
		Cc:             []string{"ops@example.com"},
		From:           "Billing <billing@example.com>",
		Subject:        "Your receipt",
		HTML:           "<p>Thanks</p>",
		Text:           "Thanks",
		ReplyTo:        "support@example.com",
		EnvelopeSender: "billing@example.com",
		Attachments:    []URLAttachment{{URL: "https://example.com/receipt.pdf", FileName: "receipt.pdf"}},
		IdempotencyKey: "order-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.Result.TrackingID != "trk_123" || resp.Data.Result.Recipients != 2 || resp.Data.Result.Status != "queued" {
		t.Fatalf("unexpected result: %+v", resp.Data.Result)
	}
	if resp.Message != "queued" {
		t.Fatalf("message = %q", resp.Message)
	}
}

func TestSpamCheckScores(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sending/spam-check" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, `{"success":true,"message":"ok","data":{"score":1.2,"threshold":5,"isSpam":false,"rating":"low_risk","badgeColor":"green","deliverabilityScore":88,"summary":"Low risk","rules":[{"code":"MISSING_UNSUBSCRIBE","points":1.2,"title":"No unsubscribe","description":"Missing","recommendation":"Add one"}]}}`)
	})
	resp, err := client.Messages.SpamCheck(context.Background(), SpamCheckRequest{Subject: "Hello", Text: "Body"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.Score != 1.2 || resp.Data.Threshold != 5 || resp.Data.IsSpam || resp.Data.DeliverabilityScore != 88 {
		t.Fatalf("unexpected report: %+v", resp.Data)
	}
	if len(resp.Data.Rules) != 1 || resp.Data.Rules[0].Code != "MISSING_UNSUBSCRIBE" {
		t.Fatalf("rules = %+v", resp.Data.Rules)
	}
}

func TestSendWithAttachments(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sending/messages/with-attachments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			t.Fatalf("content type: %q %v", r.Header.Get("Content-Type"), err)
		}
		reader := multipart.NewReader(r.Body, params["boundary"])
		form, err := reader.ReadForm(10 << 20)
		if err != nil {
			t.Fatal(err)
		}
		if len(form.Value["message"]) != 1 || !strings.Contains(form.Value["message"][0], `"subject":"Invoice"`) {
			t.Errorf("message part = %#v", form.Value["message"])
		}
		files := form.File["files"]
		if len(files) != 1 || files[0].Filename != "invoice.txt" {
			t.Fatalf("files = %#v", files)
		}
		file, _ := files[0].Open()
		data, _ := io.ReadAll(file)
		_ = file.Close()
		if string(data) != "hello attachment" {
			t.Errorf("file body = %q", data)
		}
		writeJSON(w, http.StatusOK, `{"success":true,"data":{"result":{"status":"queued","trackingId":"trk_attach","recipients":1}}}`)
	})
	resp, err := client.Messages.SendWithAttachments(context.Background(), SendMessageRequest{
		To:      []string{"ada@example.com"},
		Subject: "Invoice",
		Text:    "Attached",
	}, []FileUpload{{FileName: "invoice.txt", ContentType: "text/plain", Content: []byte("hello attachment")}})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.Result.TrackingID != "trk_attach" {
		t.Fatalf("result = %+v", resp.Data.Result)
	}
}

func TestUploadAttachment(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sending/attachments/upload" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, `{"success":true,"data":{"file":{"fileName":"logo.png","contentType":"image/png","sizeBytes":4,"publicUrl":"https://cdn.example.com/logo.png"}}}`)
	})
	resp, err := client.Messages.UploadAttachment(context.Background(), FileUpload{FileName: "logo.png", ContentType: "image/png", Content: []byte("1234")})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Data.File.FileName != "logo.png" || resp.Data.File.SizeBytes != 4 || resp.Data.File.PublicURL == nil {
		t.Fatalf("file = %+v", resp.Data.File)
	}
}
