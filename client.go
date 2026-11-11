// Package wazobiamail is the official Go SDK for the WazobiaMail API.
//
// WazobiaMail is a Resend alternative for transactional email through an API
// or SMTP, with sending domains, hosted mailboxes and deliverability tools.
// Start free with 3,000 emails/month.
package wazobiamail

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	// Version is the SDK version.
	Version = "0.1.0"
	// DefaultBaseURL is the production WazobiaMail API root.
	DefaultBaseURL = "https://app.wazobiamail.com/api/v1"
	// DefaultTimeout is the default HTTP timeout.
	DefaultTimeout = 30 * time.Second
	// DefaultUserAgent identifies the SDK to the API.
	DefaultUserAgent = "wazobiamail-go/" + Version
)

// Client is a WazobiaMail API client. It is safe for concurrent use.
type Client struct {
	baseURL     string
	apiKey      string
	accessToken string
	userAgent   string
	httpClient  *http.Client
	headers     http.Header
	allowEmpty  bool

	Messages  *MessagesService
	Domains   *DomainsService
	Mailboxes *MailboxesService
	APIKeys   *APIKeysService
	SMTP      *SMTPService
	Billing   *BillingService
	Units     *UnitsService
	Plans     *PlansService
	Status    *StatusService
	Support   *SupportService
	Templates *TemplatesService
	Webhooks  *WebhooksService
}

// Option configures a Client.
type Option func(*Client) error

// WithBaseURL overrides the API base URL.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) error {
		baseURL = strings.TrimSpace(baseURL)
		if baseURL == "" {
			return errors.New("wazobiamail: base URL must not be empty")
		}
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			return errors.New("wazobiamail: base URL must start with http:// or https://")
		}
		c.baseURL = strings.TrimRight(baseURL, "/")
		return nil
	}
}

// WithHTTPClient uses a custom *http.Client (useful for proxies, tracing and tests).
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return errors.New("wazobiamail: http client must not be nil")
		}
		c.httpClient = client
		return nil
	}
}

// WithTimeout sets the request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		if timeout <= 0 {
			return errors.New("wazobiamail: timeout must be greater than zero")
		}
		c.httpClient.Timeout = timeout
		return nil
	}
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(userAgent string) Option {
	return func(c *Client) error {
		userAgent = strings.TrimSpace(userAgent)
		if userAgent == "" {
			return errors.New("wazobiamail: user agent must not be empty")
		}
		c.userAgent = userAgent
		return nil
	}
}

// WithAccessToken authenticates with an account Bearer token instead of an API key.
func WithAccessToken(token string) Option {
	return func(c *Client) error {
		c.accessToken = strings.TrimSpace(token)
		return nil
	}
}

// WithHeader adds a header sent with every request.
func WithHeader(name, value string) Option {
	return func(c *Client) error {
		if strings.TrimSpace(name) == "" {
			return errors.New("wazobiamail: header name must not be empty")
		}
		c.headers.Set(name, value)
		return nil
	}
}

// WithAllowUnauthenticated permits construction without a key for public
// endpoints such as plans and status.
func WithAllowUnauthenticated(allow bool) Option {
	return func(c *Client) error {
		c.allowEmpty = allow
		return nil
	}
}

// NewClient creates a client. An empty apiKey falls back to WAZOBIA_API_KEY or
// WAZOBIAMAIL_API_KEY.
func NewClient(apiKey string, options ...Option) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		apiKey = firstEnv("WAZOBIA_API_KEY", "WAZOBIAMAIL_API_KEY")
	}

	client := &Client{
		baseURL:    DefaultBaseURL,
		apiKey:     apiKey,
		userAgent:  DefaultUserAgent,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		headers:    make(http.Header),
	}
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(client); err != nil {
			return nil, err
		}
	}
	if client.apiKey == "" && client.accessToken == "" && !client.allowEmpty {
		return nil, errors.New("wazobiamail: an API key or access token is required")
	}

	client.Messages = &MessagesService{client: client}
	client.Domains = &DomainsService{client: client}
	client.Mailboxes = &MailboxesService{client: client}
	client.APIKeys = &APIKeysService{client: client}
	client.SMTP = &SMTPService{client: client}
	client.Billing = &BillingService{client: client}
	client.Units = &UnitsService{client: client}
	client.Plans = &PlansService{client: client}
	client.Status = &StatusService{client: client}
	client.Support = &SupportService{client: client}
	client.Templates = &TemplatesService{client: client}
	client.Webhooks = &WebhooksService{}
	return client, nil
}

// NewClientFromEnv creates a client from WAZOBIA_API_KEY or WAZOBIAMAIL_API_KEY.
func NewClientFromEnv(options ...Option) (*Client, error) {
	return NewClient("", options...)
}

// BaseURL returns the configured API root.
func (c *Client) BaseURL() string { return c.baseURL }

// String returns a log-safe representation of the client.
func (c *Client) String() string {
	return "WazobiaMail{baseURL:" + c.baseURL + ", apiKey:" + maskSecret(c.apiKey) + "}"
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func maskSecret(secret string) string {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "***"
	}
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}
