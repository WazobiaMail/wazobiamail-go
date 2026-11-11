# WazobiaMail Go SDK

The official Go client for [WazobiaMail](https://wazobiamail.com) — transactional
email through an **API or SMTP**, with sending domains, hosted mailboxes and
deliverability tools.

Start free with **3,000 emails/month**.

[![Go Reference](https://pkg.go.dev/badge/github.com/WazobiaMail/wazobiamail-go.svg)](https://pkg.go.dev/github.com/WazobiaMail/wazobiamail-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/WazobiaMail/wazobiamail-go)](https://goreportcard.com/report/github.com/WazobiaMail/wazobiamail-go)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

WazobiaMail is a **Resend alternative** for Go applications: send through the
REST API or SMTP, manage verified domains and DNS, provision hosted mailboxes,
and inspect delivery, open and click activity.

## Install

```bash
go get github.com/WazobiaMail/wazobiamail-go@v0.1.0
```

```go
import wazobiamail "github.com/WazobiaMail/wazobiamail-go"
```

Requires Go 1.22+ and has **no third-party dependencies**.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	wazobiamail "github.com/WazobiaMail/wazobiamail-go"
)

func main() {
	client, err := wazobiamail.NewClient("wz_live_your_api_key")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.Messages.Send(context.Background(), wazobiamail.SendMessageRequest{
		To:      []string{"customer@example.com"},
		From:    "Billing <billing@example.com>",
		Subject: "Your receipt",
		HTML:    "<p>Thanks for your order.</p>",
		Text:    "Thanks for your order.",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Data.Result.TrackingID)
}
```

`NewClient` reads `WAZOBIA_API_KEY` or `WAZOBIAMAIL_API_KEY` when the key is
empty, so you can also call `wazobiamail.NewClientFromEnv()`.

## Configuration

```go
client, err := wazobiamail.NewClient(
	"wz_live_your_api_key",
	wazobiamail.WithTimeout(15*time.Second),
	wazobiamail.WithUserAgent("my-app/2.0"),
	wazobiamail.WithHeader("X-Trace-Id", requestID),
)
```

| Option | Purpose |
| --- | --- |
| `WithBaseURL` | Override `https://app.wazobiamail.com/api/v1` (staging/proxy/tests). |
| `WithHTTPClient` | Bring your own `*http.Client` for tracing, proxies or mocks. |
| `WithTimeout` | Request timeout; default `30s`. |
| `WithAccessToken` | Authenticate with an account Bearer token instead of `X-API-Key`. |
| `WithHeader` | Add a header to every request. |
| `WithAllowUnauthenticated` | Construct a client for public endpoints without a key. |

The `*Client` is safe for concurrent use. Pass a `context.Context` to every
method for cancellation and deadlines.

## Sending email

`SendMessageRequest` is flat and ergonomic; the SDK builds the canonical nested
payload for you.

```go
resp, err := client.Messages.Send(ctx, wazobiamail.SendMessageRequest{
	To:      []string{"ada@example.com", "ops@example.com"},
	Cc:      []string{"finance@example.com"},
	Bcc:     []string{"archive@example.com"},
	From:    "Billing <billing@example.com>",
	Subject: "Invoice INV-2026-001",
	HTML:    "<h1>Invoice</h1><p>Your invoice is attached.</p>",
	Text:    "Invoice attached.",
	ReplyTo: "support@example.com",
	Attachments: []wazobiamail.URLAttachment{
		{URL: "https://cdn.example.com/invoice.pdf", FileName: "invoice.pdf"},
	},
	IdempotencyKey: "invoice-INV-2026-001",
})
```

Idempotency keys prevent duplicate sends when a request is retried. The API
stores the receipt in `resp.Data.Result` (`TrackingID`, `Status`, `Recipients`).

### Binary attachments

```go
resp, err := client.Messages.SendWithAttachments(ctx, message,
	[]wazobiamail.FileUpload{
		{FileName: "invoice.pdf", ContentType: "application/pdf", Content: pdfBytes},
	},
)
```

Or upload first and use the public URL:

```go
upload, err := client.Messages.UploadAttachment(ctx, wazobiamail.FileUpload{
	FileName: "logo.png",
	ContentType: "image/png",
	Content: logoBytes,
})
fmt.Println(upload.Data.File.PublicURL)
```

## Reading spam-check scores

`Messages.SpamCheck` returns a penalty `score`, a `threshold`, an `isSpam`
boolean and an inverse `deliverabilityScore` (0–100, higher is better).

| `score` | `rating` | `isSpam` | `deliverabilityScore` | Decision |
| --- | --- | --- | --- | --- |
| `0.0 – 1.9` | `low_risk` | `false` | `81 – 100` | **Okay to send.** Clean or only minor findings. |
| `2.0 – 4.9` | `moderate_risk` | `false` | `51 – 80` | **Review first.** Fix the highest-point rules before sending. |
| `5.0 +` | `high_risk` | `true` | `0 – 50` | **Do not send.** Fix the findings and re-run the check. |

`IsSpam` is authoritative: it is true when `Score >= Threshold`. Read
`Threshold` from the response instead of hard-coding `5.0`. A green score is not
a delivery guarantee.

```go
report, err := client.Messages.SpamCheck(ctx, wazobiamail.SpamCheckRequest{
	Subject: "Your account summary",
	HTML:    "<p>Your summary is ready.</p>",
	Text:    "Your summary is ready.",
})

if report.Data.IsSpam {
	for _, rule := range report.Data.Rules {
		log.Printf("%s (+%.1f): %s", rule.Code, rule.Points, rule.Recommendation)
	}
}
```

Full guide: [Spam and deliverability scores](https://docs.wazobiamail.com/docs/guides/spam-check).

## Sending domains and DNS

```go
created, err := client.Domains.Create(ctx, wazobiamail.CreateDomainRequest{
	Domain: "example.com",
})

dns, err := client.Domains.DNSRecords(ctx, created.Data.Domain.ID)
for _, record := range dns.Data.Domain.Records {
	fmt.Printf("%s %s %s\n", record.RecordType, record.Host, record.Value)
}

verified, err := client.Domains.Verify(ctx, created.Data.Domain.ID, true)
fmt.Println(verified.Data.Verification.VerificationStatus)
```

Available: `List`, `Create`, `Get`, `Update`, `Delete`, `DNSRecords`, `Verify`,
`RefreshDNS`, and Cloudflare `CloudflareStatus`, `CloudflareConnect`,
`CloudflareSync`, `CloudflareDisconnect`.

## Hosted mailboxes

```go
mailbox, err := client.Mailboxes.Create(ctx, wazobiamail.CreateMailboxRequest{
	DomainID:    domainID,
	LocalPart:   "ada",
	DisplayName: "Ada Lovelace",
})

connection, err := client.Mailboxes.Connection(ctx, mailbox.Data.ID)
fmt.Println(connection.Data.SMTP["host"])

messages, err := client.Mailboxes.ListMessages(ctx, mailbox.Data.ID,
	wazobiamail.ListMessagesParams{UnreadOnly: true, Limit: 20})
```

Also available: `List`, `Get`, `Update`, `Delete`, `CreateCredential`,
`RotateCredential`, `RevokeCredential`, `Storage`, `GetMessage`, `DeleteMessage`.

## API keys, SMTP relays and access

```go
key, err := client.APIKeys.Create(ctx, wazobiamail.CreateAPIKeyRequest{
	Name:  "production",
	Scope: "full",
})
fmt.Println(key.Data.APIKey.APIKey) // one-time secret

relay, err := client.SMTP.CreateRelay(ctx, wazobiamail.CreateSMTPRelayRequest{
	Name:         "primary",
	SMTPHost:     "smtp.example.com",
	SMTPPort:     587,
	SMTPUsername: "user",
	SMTPPassword: "password",
	SMTPSecurity: "starttls",
})

access, err := client.SMTP.Access(ctx)
fmt.Println(access.Data.Enabled)
```

## Billing, units, plans and status

```go
invoices, err := client.Billing.ListInvoices(ctx)
pdf, headers, err := client.Billing.DownloadInvoice(ctx, invoiceID)
subscription, err := client.Billing.CurrentSubscription(ctx, "email")

summary, err := client.Units.Summary(ctx)
packs, err := client.Units.Packs(ctx)
ledger, err := client.Units.Ledger(ctx, 1, 25)

plans, err := client.Plans.List(ctx)
limits, err := client.Plans.EffectiveLimits(ctx)
status, err := client.Status.Get(ctx)
```

## Templates and support

```go
template, err := client.Templates.Create(ctx, wazobiamail.CreateTemplateRequest{
	Name:    "Welcome",
	Subject: "Welcome, {{name}}",
	Body:    "<p>Hi {{name}}</p>",
})

preview, err := client.Templates.Preview(ctx, template.Data.ID,
	wazobiamail.PreviewTemplateRequest{SampleVariables: map[string]any{"name": "Ada"}})

ticket, err := client.Support.Create(ctx, wazobiamail.CreateSupportTicketRequest{
	Name:     "Ada",
	Email:    "ada@example.com",
	Subject:  "DNS help",
	Category: "deliverability",
	Priority: "normal",
	Message:  "My domain is still pending.",
})
```

Templates require Enterprise access (or an admin account). Support tickets can
include base64 `SupportUpload` attachments up to 5 MiB each.

## Webhooks

Verify the raw request body before decoding it:

```go
signature := r.Header.Get(wazobiamail.WebhookSignatureHeader)

var event struct {
	Type string `json:"type"`
	Data struct {
		TrackingID string `json:"trackingId"`
	} `json:"data"`
}

if err := client.Webhooks.ConstructEvent(body, signature, secret, &event); err != nil {
	if errors.Is(err, wazobiamail.ErrInvalidWebhookSignature) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	http.Error(w, "bad payload", http.StatusBadRequest)
	return
}
```

`VerifySignature` accepts the hex HMAC SHA-256 value and an optional
`sha256=` prefix. `Sign` is useful in tests.

## Errors

```go
var apiErr *wazobiamail.APIError
if errors.As(err, &apiErr) {
	switch {
	case apiErr.IsValidation():
		fmt.Println(apiErr.Errors)
	case apiErr.IsRateLimited():
		time.Sleep(apiErr.RetryAfter)
	case apiErr.IsNotFound():
		// ...
	}
}
```

`APIError` exposes `StatusCode`, `Message`, `Code`, `Errors`, `Body` and
`RetryAfter`. `IsUnauthorized`, `IsForbidden`, `IsValidation`, `IsNotFound` and
`IsRateLimited` are convenience helpers.

## Feature summary

| Service | Methods |
| --- | --- |
| `Messages` | `Send`, `SendWithAttachments`, `UploadAttachment`, `SpamCheck`, `ListTracked`, `GetTracked`, `ListEmailLogs` |
| `Domains` | `List`, `Create`, `Get`, `Update`, `Delete`, `DNSRecords`, `Verify`, `RefreshDNS`, Cloudflare connect/sync/status/disconnect |
| `Mailboxes` | `List`, `Create`, `Get`, `Update`, `Delete`, credentials, `Storage`, messages |
| `APIKeys` | `List`, `Create`, `Regenerate`, `Revoke` |
| `SMTP` | relays CRUD/activate/deactivate, `Access`, `UpdateAccess` |
| `Billing` | `ListInvoices`, `DownloadInvoice`, `CurrentSubscription` |
| `Units` | `Summary`, `Packs`, `Ledger`, `Purchase` |
| `Plans` | `List`, `EffectiveLimits` |
| `Status` | `Get` |
| `Support` | `Create`, `List`, `Get`, `Reply`, `MarkRead`, `DownloadAttachment` |
| `Templates` | `List`, `Create`, `Get`, `Update`, `Delete`, `Preview`, `SyncMeta` |
| `Webhooks` | `VerifySignature`, `ConstructEvent`, `Sign` |

## Development

```bash
git clone https://github.com/WazobiaMail/wazobiamail-go.git
cd wazobiamail-go
go test -race -cover ./...
go vet ./...
gofmt -l .
```

## Links

- [Documentation](https://docs.wazobiamail.com/docs/guides/sdks/go)
- [API reference](https://docs.wazobiamail.com/docs/api-reference)
- [pkg.go.dev](https://pkg.go.dev/github.com/WazobiaMail/wazobiamail-go)
- [GitHub](https://github.com/WazobiaMail/wazobiamail-go)
- [WazobiaMail](https://wazobiamail.com)
- [Terms](https://wazobiamail.com/terms/) · [Privacy](https://wazobiamail.com/privacy/)

## License

MIT © Bonifade Technologies
