package wazobiamail

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDomains(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/sending/domains":
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"items":[{"id":"dom_1","domain":"example.com","verificationStatus":"verified"}]}}`)
		case r.Method == http.MethodPost && r.URL.Path == "/sending/domains":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["domain"] != "example.com" {
				t.Errorf("domain body = %#v", body)
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"domain":{"id":"dom_1","domain":"example.com","verificationStatus":"pending"}}}`)
		case r.URL.Path == "/sending/domains/dom_1/dns-records":
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"domain":{"domainId":"dom_1","domain":"example.com","provisioning":false,"records":[{"recordType":"TXT","host":"@","value":"v=spf1","status":"pending"}]}}}`)
		case r.URL.Path == "/sending/domains/dom_1/verify":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["force"] != true {
				t.Errorf("force = %#v", body["force"])
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"verification":{"domainId":"dom_1","domain":"example.com","verificationStatus":"verified","verifiedCount":1,"pendingCount":0}}}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
		}
	})

	list, err := client.Domains.List(context.Background())
	if err != nil || len(list.Data.Items) != 1 || list.Data.Items[0].ID != "dom_1" {
		t.Fatalf("list: %+v %v", list, err)
	}
	created, err := client.Domains.Create(context.Background(), CreateDomainRequest{Domain: "example.com"})
	if err != nil || created.Data.Domain.VerificationStatus != "pending" {
		t.Fatalf("create: %+v %v", created, err)
	}
	dns, err := client.Domains.DNSRecords(context.Background(), "dom_1")
	if err != nil || len(dns.Data.Domain.Records) != 1 || dns.Data.Domain.Records[0].RecordType != "TXT" {
		t.Fatalf("dns: %+v %v", dns, err)
	}
	verification, err := client.Domains.Verify(context.Background(), "dom_1", true)
	if err != nil || verification.Data.Verification.VerifiedCount != 1 {
		t.Fatalf("verify: %+v %v", verification, err)
	}
}

func TestMailboxes(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/mailboxes":
			writeJSON(w, http.StatusOK, `{"success":true,"data":[{"id":"mb_1","address":"ada@example.com","status":"active"}]}`)
		case r.Method == http.MethodPost && r.URL.Path == "/mailboxes":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["localPart"] != "ada" || body["domainId"] != "dom_1" {
				t.Errorf("mailbox body = %#v", body)
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"id":"mb_1","address":"ada@example.com","status":"active"}}`)
		case r.URL.Path == "/mailboxes/mb_1/messages":
			if got := r.URL.Query().Get("unreadOnly"); got != "true" {
				t.Errorf("unreadOnly = %q", got)
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"items":[{"id":"msg_1","subject":"Hello"}],"total":1,"page":1,"limit":20}}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
		}
	})
	list, err := client.Mailboxes.List(context.Background())
	if err != nil || len(list.Data) != 1 || list.Data[0].Address != "ada@example.com" {
		t.Fatalf("list: %+v %v", list, err)
	}
	created, err := client.Mailboxes.Create(context.Background(), CreateMailboxRequest{DomainID: "dom_1", LocalPart: "ada"})
	if err != nil || created.Data.ID != "mb_1" {
		t.Fatalf("create: %+v %v", created, err)
	}
	messages, err := client.Mailboxes.ListMessages(context.Background(), "mb_1", ListMessagesParams{UnreadOnly: true})
	if err != nil || messages.Data.Total != 1 || messages.Data.Items[0].Subject == nil {
		t.Fatalf("messages: %+v %v", messages, err)
	}
}

func TestAPIKeysAndSMTP(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sending/api-keys":
			if r.Method != http.MethodPost {
				t.Errorf("method = %s", r.Method)
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"apiKey":{"id":"key_1","name":"prod","scope":"full","keyPrefix":"wz_live","apiKey":"wz_live_secret"}}}`)
		case "/smtp-relays":
			if r.Method != http.MethodPost {
				t.Errorf("method = %s", r.Method)
			}
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["smtp_host"] != "smtp.example.com" || body["smtp_port"] != float64(587) {
				t.Errorf("relay body = %#v", body)
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"smtpRelay":{"id":"relay_1","name":"primary","status":"active"}}}`)
		case "/sending/smtp-access":
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"paid":true,"eligible":true,"enabled":true,"configured":true}}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
		}
	})
	key, err := client.APIKeys.Create(context.Background(), CreateAPIKeyRequest{Name: "prod", Scope: "full"})
	if err != nil || key.Data.APIKey.APIKey != "wz_live_secret" {
		t.Fatalf("api key: %+v %v", key, err)
	}
	relay, err := client.SMTP.CreateRelay(context.Background(), CreateSMTPRelayRequest{Name: "primary", SMTPHost: "smtp.example.com", SMTPPort: 587, SMTPUsername: "user", SMTPPassword: "pass", SMTPSecurity: "starttls"})
	if err != nil || relay.Data.SMTPRelay.ID != "relay_1" {
		t.Fatalf("relay: %+v %v", relay, err)
	}
	access, err := client.SMTP.Access(context.Background())
	if err != nil || !access.Data.Enabled {
		t.Fatalf("access: %+v %v", access, err)
	}
}

func TestTemplatesAndSupport(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/templates":
			if r.Method != http.MethodGet {
				t.Errorf("method = %s", r.Method)
			}
			if got := r.URL.Query().Get("channel"); got != "email" {
				t.Errorf("channel = %q", got)
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"items":[{"id":"tpl_1","name":"Welcome","body":"Hi {{name}}"}],"total":1,"limit":50,"offset":0}}`)
		case "/templates/tpl_1/preview":
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"channel":"email","subject":"Welcome","renderedBody":"Hi Ada","renderedText":"Hi Ada","status":"active","canSend":true}}`)
		case "/support":
			writeJSON(w, http.StatusCreated, `{"success":true,"data":{"id":"ticket_1","ticketId":"ticket_1"}}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
		}
	})
	list, err := client.Templates.List(context.Background(), TemplateListParams{Channel: "email"})
	if err != nil || len(list.Data.Items) != 1 || list.Data.Items[0].Name != "Welcome" {
		t.Fatalf("templates: %+v %v", list, err)
	}
	preview, err := client.Templates.Preview(context.Background(), "tpl_1", PreviewTemplateRequest{SampleVariables: map[string]any{"name": "Ada"}})
	if err != nil || preview.Data.RenderedBody != "Hi Ada" || !preview.Data.CanSend {
		t.Fatalf("preview: %+v %v", preview, err)
	}
	ticket, err := client.Support.Create(context.Background(), CreateSupportTicketRequest{Name: "Ada", Email: "ada@example.com", Subject: "Help", Category: "other", Priority: "normal", Message: "Hello"})
	if err != nil || ticket.Data.TicketID != "ticket_1" {
		t.Fatalf("ticket: %+v %v", ticket, err)
	}
}

func TestBillingUnitsPlansStatus(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/billing/invoices/inv_1/download":
			w.Header().Set("Content-Type", "application/pdf")
			_, _ = w.Write([]byte("%PDF-test"))
		case "/subscriptions/current":
			if got := r.URL.Query().Get("product"); got != "email" {
				t.Errorf("product = %q", got)
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"subscription":{"id":"sub_1","planId":"business","status":"active"}}}`)
		case "/units/summary":
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"includedAllowance":3000,"includedRemaining":2500,"purchasedBalance":100,"unlimited":false}}`)
		case "/units/ledger":
			if got := r.URL.Query().Get("per_page"); got != "25" {
				t.Errorf("per_page = %q", got)
			}
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"items":[{"id":"led_1","delta":-2,"balanceAfter":2500}],"pagination":{"page":1}}}`)
		case "/plans":
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"items":[{"id":"free","name":"Free","monthlyQuota":3000}],"pagination":{"page":1}}}`)
		case "/plans/effective":
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"limits":{"accountId":1,"planId":"free","sendRateLimitPerMinute":30}}}`)
		case "/status":
			writeJSON(w, http.StatusOK, `{"success":true,"data":{"plan":{"id":"free"},"units":{"remaining":2500}}}`)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			writeJSON(w, http.StatusNotFound, `{"message":"not found"}`)
		}
	})
	if body, _, err := client.Billing.DownloadInvoice(context.Background(), "inv_1"); err != nil || string(body) != "%PDF-test" {
		t.Fatalf("invoice: %q %v", body, err)
	}
	subscription, err := client.Billing.CurrentSubscription(context.Background(), "email")
	if err != nil || subscription.Data.Subscription == nil || subscription.Data.Subscription.ID != "sub_1" {
		t.Fatalf("subscription: %+v %v", subscription, err)
	}
	summary, err := client.Units.Summary(context.Background())
	if err != nil || summary.Data.IncludedAllowance != 3000 {
		t.Fatalf("units: %+v %v", summary, err)
	}
	ledger, err := client.Units.Ledger(context.Background(), 1, 25)
	if err != nil || len(ledger.Data.Items) != 1 {
		t.Fatalf("ledger: %+v %v", ledger, err)
	}
	plans, err := client.Plans.List(context.Background())
	if err != nil || plans.Data.Items[0].MonthlyQuota != 3000 {
		t.Fatalf("plans: %+v %v", plans, err)
	}
	limits, err := client.Plans.EffectiveLimits(context.Background())
	if err != nil || limits.Data.Limits.SendRateLimitPerMinute != 30 {
		t.Fatalf("limits: %+v %v", limits, err)
	}
	status, err := client.Status.Get(context.Background())
	if err != nil || status.Data.Plan == nil {
		t.Fatalf("status: %+v %v", status, err)
	}
}
