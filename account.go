package wazobiamail

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// APIKey is a scoped sending API key.
type APIKey struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Scope      string  `json:"scope"`
	DomainID   *string `json:"domainId"`
	Domain     *string `json:"domain"`
	KeyPrefix  string  `json:"keyPrefix"`
	LastUsedAt *string `json:"lastUsedAt"`
	ExpiresAt  *string `json:"expiresAt"`
	RevokedAt  *string `json:"revokedAt"`
	CreatedAt  string  `json:"createdAt"`
}

// CreateAPIKeyRequest creates a sending key.
type CreateAPIKeyRequest struct {
	Name      string `json:"name"`
	Scope     string `json:"scope,omitempty"`
	DomainID  string `json:"domainId,omitempty"`
	ExpiresAt string `json:"expiresAt,omitempty"`
}

// CreatedAPIKey includes the one-time secret.
type CreatedAPIKey struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Scope     string  `json:"scope"`
	DomainID  *string `json:"domainId"`
	KeyPrefix string  `json:"keyPrefix"`
	APIKey    string  `json:"apiKey"`
	ExpiresAt *string `json:"expiresAt"`
}

// APIKeyEnvelope is {"apiKey":{...}}.
type APIKeyEnvelope struct {
	APIKey CreatedAPIKey `json:"apiKey"`
}

// APIKeysService manages sending API keys.
type APIKeysService struct{ client *Client }

func (s *APIKeysService) List(ctx context.Context) (*APIResponse[Items[APIKey]], error) {
	return doJSON[Items[APIKey]](ctx, s.client, http.MethodGet, "/sending/api-keys", nil, requestOptions{})
}

func (s *APIKeysService) Create(ctx context.Context, req CreateAPIKeyRequest) (*APIResponse[APIKeyEnvelope], error) {
	return doJSON[APIKeyEnvelope](ctx, s.client, http.MethodPost, "/sending/api-keys", req, requestOptions{})
}

func (s *APIKeysService) Regenerate(ctx context.Context, keyID string) (*APIResponse[APIKeyEnvelope], error) {
	return doJSON[APIKeyEnvelope](ctx, s.client, http.MethodPost, "/sending/api-keys/"+url.PathEscape(keyID)+"/regenerate", map[string]any{}, requestOptions{})
}

func (s *APIKeysService) Revoke(ctx context.Context, keyID string) (*APIResponse[map[string]string], error) {
	return doJSON[map[string]string](ctx, s.client, http.MethodPost, "/sending/api-keys/"+url.PathEscape(keyID)+"/revoke", map[string]any{}, requestOptions{})
}

// SMTPRelay is a customer SMTP relay configuration.
type SMTPRelay struct {
	ID           string `json:"id"`
	AccountID    int64  `json:"accountId"`
	Name         string `json:"name"`
	SMTPHost     string `json:"smtpHost"`
	SMTPPort     int    `json:"smtpPort"`
	SMTPUsername string `json:"smtpUsername"`
	SMTPSecurity string `json:"smtpSecurity"`
	Status       string `json:"status"`
	IsDefault    bool   `json:"isDefault"`
	HasPassword  bool   `json:"hasPassword"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// SMTPRelayEnvelope is {"smtpRelay":{...}}.
type SMTPRelayEnvelope struct {
	SMTPRelay SMTPRelay `json:"smtpRelay"`
}

// SMTPRelaysEnvelope is {"smtpRelays":[...]}.
type SMTPRelaysEnvelope struct {
	SMTPRelays []SMTPRelay `json:"smtpRelays"`
}

// CreateSMTPRelayRequest creates a relay. The request DTO uses snake_case.
type CreateSMTPRelayRequest struct {
	Name         string `json:"name"`
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"`
	SMTPSecurity string `json:"smtp_security"`
	IsDefault    *bool  `json:"is_default,omitempty"`
	Status       string `json:"status,omitempty"`
}

// UpdateSMTPRelayRequest updates a relay.
type UpdateSMTPRelayRequest struct {
	Name         string `json:"name,omitempty"`
	SMTPHost     string `json:"smtp_host,omitempty"`
	SMTPPort     int    `json:"smtp_port,omitempty"`
	SMTPUsername string `json:"smtp_username,omitempty"`
	SMTPPassword string `json:"smtp_password,omitempty"`
	SMTPSecurity string `json:"smtp_security,omitempty"`
	IsDefault    *bool  `json:"is_default,omitempty"`
	Status       string `json:"status,omitempty"`
}

// SMTPConnection is the SMTP access connection block.
type SMTPConnection struct {
	Host      string   `json:"host"`
	Port      int      `json:"port"`
	Security  string   `json:"security"`
	Usernames []string `json:"usernames"`
}

// SMTPAccess is the account SMTP access snapshot.
type SMTPAccess struct {
	Paid       bool            `json:"paid"`
	Eligible   bool            `json:"eligible"`
	Enabled    bool            `json:"enabled"`
	Persona    *string         `json:"persona"`
	Configured bool            `json:"configured"`
	Connection *SMTPConnection `json:"connection"`
}

// SMTPService groups relay and access management.
type SMTPService struct{ client *Client }

func (s *SMTPService) ListRelays(ctx context.Context) (*APIResponse[SMTPRelaysEnvelope], error) {
	return doJSON[SMTPRelaysEnvelope](ctx, s.client, http.MethodGet, "/smtp-relays", nil, requestOptions{})
}

func (s *SMTPService) CreateRelay(ctx context.Context, req CreateSMTPRelayRequest) (*APIResponse[SMTPRelayEnvelope], error) {
	return doJSON[SMTPRelayEnvelope](ctx, s.client, http.MethodPost, "/smtp-relays", req, requestOptions{})
}

func (s *SMTPService) GetRelay(ctx context.Context, relayID string) (*APIResponse[SMTPRelayEnvelope], error) {
	return doJSON[SMTPRelayEnvelope](ctx, s.client, http.MethodGet, "/smtp-relays/"+url.PathEscape(relayID), nil, requestOptions{})
}

func (s *SMTPService) UpdateRelay(ctx context.Context, relayID string, req UpdateSMTPRelayRequest) (*APIResponse[SMTPRelayEnvelope], error) {
	return doJSON[SMTPRelayEnvelope](ctx, s.client, http.MethodPatch, "/smtp-relays/"+url.PathEscape(relayID), req, requestOptions{})
}

func (s *SMTPService) DeleteRelay(ctx context.Context, relayID string) (*APIResponse[map[string]string], error) {
	return doJSON[map[string]string](ctx, s.client, http.MethodDelete, "/smtp-relays/"+url.PathEscape(relayID), nil, requestOptions{})
}

func (s *SMTPService) ActivateRelay(ctx context.Context, relayID string) (*APIResponse[SMTPRelayEnvelope], error) {
	return doJSON[SMTPRelayEnvelope](ctx, s.client, http.MethodPost, "/smtp-relays/"+url.PathEscape(relayID)+"/activate", map[string]any{}, requestOptions{})
}

func (s *SMTPService) DeactivateRelay(ctx context.Context, relayID string) (*APIResponse[SMTPRelayEnvelope], error) {
	return doJSON[SMTPRelayEnvelope](ctx, s.client, http.MethodPost, "/smtp-relays/"+url.PathEscape(relayID)+"/deactivate", map[string]any{}, requestOptions{})
}

func (s *SMTPService) Access(ctx context.Context) (*APIResponse[SMTPAccess], error) {
	return doJSON[SMTPAccess](ctx, s.client, http.MethodGet, "/sending/smtp-access", nil, requestOptions{})
}

func (s *SMTPService) UpdateAccess(ctx context.Context, enabled bool) (*APIResponse[map[string]bool], error) {
	return doJSON[map[string]bool](ctx, s.client, http.MethodPut, "/sending/smtp-access", map[string]bool{"enabled": enabled}, requestOptions{})
}

// Invoice is a billing invoice.
type Invoice struct {
	ID        string         `json:"id"`
	AccountID int64          `json:"accountId"`
	Amount    float64        `json:"amount"`
	Currency  string         `json:"currency"`
	Status    string         `json:"status"`
	Purpose   string         `json:"purpose"`
	Metadata  map[string]any `json:"metadata"`
	DueDate   *string        `json:"dueDate"`
	PaidAt    *string        `json:"paidAt"`
	CreatedAt string         `json:"createdAt"`
}

// Subscription is the current subscription.
type Subscription struct {
	ID              string         `json:"id"`
	AccountID       int64          `json:"accountId"`
	PlanID          string         `json:"planId"`
	Status          string         `json:"status"`
	StartDate       string         `json:"startDate"`
	EndDate         *string        `json:"endDate"`
	AutoRenew       bool           `json:"autoRenew"`
	BillingInterval string         `json:"billingInterval"`
	Plan            map[string]any `json:"plan"`
}

// SubscriptionEnvelope is {"subscription":{...}|null}.
type SubscriptionEnvelope struct {
	Subscription *Subscription `json:"subscription"`
}

// UnitsSummary is the email credit summary.
type UnitsSummary struct {
	IncludedAllowance int64 `json:"includedAllowance"`
	IncludedUsed      int64 `json:"includedUsed"`
	IncludedRemaining int64 `json:"includedRemaining"`
	PurchasedBalance  int64 `json:"purchasedBalance"`
	Unlimited         bool  `json:"unlimited"`
	CanBuyCredits     bool  `json:"canBuyCredits"`
}

// UnitPack is a purchasable credit pack.
type UnitPack struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Units         int64  `json:"units"`
	PriceNGNMinor int64  `json:"priceNgnMinor"`
	PriceUSDMinor int64  `json:"priceUsdMinor"`
	IsActive      bool   `json:"isActive"`
	SortOrder     int    `json:"sortOrder"`
}

// UnitLedgerEntry is a credit ledger row.
type UnitLedgerEntry struct {
	ID           string  `json:"id"`
	Delta        int64   `json:"delta"`
	BalanceAfter int64   `json:"balanceAfter"`
	Reason       string  `json:"reason"`
	Reference    *string `json:"reference"`
	CreatedAt    string  `json:"createdAt"`
}

// PurchaseUnitsRequest buys a credit pack.
type PurchaseUnitsRequest struct {
	PackID   string `json:"packId"`
	Currency string `json:"currency"`
}

// BillingService reads invoices and subscriptions.
type BillingService struct{ client *Client }

func (s *BillingService) ListInvoices(ctx context.Context) (*APIResponse[Page[Invoice]], error) {
	return doJSON[Page[Invoice]](ctx, s.client, http.MethodGet, "/billing/invoices", nil, requestOptions{})
}

func (s *BillingService) DownloadInvoice(ctx context.Context, invoiceID string) ([]byte, http.Header, error) {
	return s.client.doBytes(ctx, "/billing/invoices/"+url.PathEscape(invoiceID)+"/download", requestOptions{})
}

func (s *BillingService) CurrentSubscription(ctx context.Context, product string) (*APIResponse[SubscriptionEnvelope], error) {
	opts := requestOptions{}
	if product != "" {
		opts.query = url.Values{"product": []string{product}}
	}
	return doJSON[SubscriptionEnvelope](ctx, s.client, http.MethodGet, "/subscriptions/current", nil, opts)
}

// UnitsService reads and purchases email credits.
type UnitsService struct{ client *Client }

func (s *UnitsService) Summary(ctx context.Context) (*APIResponse[UnitsSummary], error) {
	return doJSON[UnitsSummary](ctx, s.client, http.MethodGet, "/units/summary", nil, requestOptions{})
}

func (s *UnitsService) Packs(ctx context.Context) (*APIResponse[Items[UnitPack]], error) {
	return doJSON[Items[UnitPack]](ctx, s.client, http.MethodGet, "/units/packs", nil, requestOptions{})
}

func (s *UnitsService) Ledger(ctx context.Context, page, perPage int) (*APIResponse[Page[UnitLedgerEntry]], error) {
	query := url.Values{}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if perPage > 0 {
		query.Set("per_page", strconv.Itoa(perPage))
	}
	return doJSON[Page[UnitLedgerEntry]](ctx, s.client, http.MethodGet, "/units/ledger", nil, requestOptions{query: query})
}

func (s *UnitsService) Purchase(ctx context.Context, req PurchaseUnitsRequest) (*APIResponse[map[string]any], error) {
	return doJSON[map[string]any](ctx, s.client, http.MethodPost, "/units/purchases", req, requestOptions{})
}

// Plan is a subscription plan.
type Plan struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	USDPrice               float64  `json:"usdPrice"`
	NairaPrice             float64  `json:"nairaPrice"`
	YearlyUSDPrice         *float64 `json:"yearlyUsdPrice"`
	YearlyNairaPrice       *float64 `json:"yearlyNairaPrice"`
	HourlyQuota            int      `json:"hourlyQuota"`
	DailyQuota             int      `json:"dailyQuota"`
	MonthlyQuota           int      `json:"monthlyQuota"`
	SendRateLimitPerMinute int      `json:"sendRateLimitPerMinute"`
	DataRetentionDays      int      `json:"dataRetentionDays"`
	Features               any      `json:"features"`
	Limits                 any      `json:"limits"`
	Status                 string   `json:"status"`
}

// EffectiveLimits is the resolved account limits.
type EffectiveLimits struct {
	AccountID                  int64  `json:"accountId"`
	PlanID                     string `json:"planId"`
	SendRateLimitPerMinute     int    `json:"sendRateLimitPerMinute"`
	MailboxOpsPerMinute        int    `json:"mailboxOpsPerMinute"`
	MailboxLimit               int    `json:"mailboxLimit"`
	BaseMailboxLimit           int    `json:"baseMailboxLimit"`
	AdditionalMailboxLimit     int    `json:"additionalMailboxLimit"`
	CustomDomainSendingEnabled bool   `json:"customDomainSendingEnabled"`
	DomainLimit                int    `json:"domainLimit"`
	ReceivingEnabled           bool   `json:"receivingEnabled"`
	MailboxStorageGB           int    `json:"mailboxStorageGb"`
}

// LimitsEnvelope is {"limits":{...}}.
type LimitsEnvelope struct {
	Limits EffectiveLimits `json:"limits"`
}

// PlansService reads plans and limits.
type PlansService struct{ client *Client }

func (s *PlansService) List(ctx context.Context) (*APIResponse[Page[Plan]], error) {
	return doJSON[Page[Plan]](ctx, s.client, http.MethodGet, "/plans", nil, requestOptions{})
}

func (s *PlansService) EffectiveLimits(ctx context.Context) (*APIResponse[LimitsEnvelope], error) {
	return doJSON[LimitsEnvelope](ctx, s.client, http.MethodGet, "/plans/effective", nil, requestOptions{})
}

// AccountStatus is the /status response.
type AccountStatus struct {
	Plan  map[string]any `json:"plan"`
	Units map[string]any `json:"units"`
}

// StatusService reads account health.
type StatusService struct{ client *Client }

func (s *StatusService) Get(ctx context.Context) (*APIResponse[AccountStatus], error) {
	return doJSON[AccountStatus](ctx, s.client, http.MethodGet, "/status", nil, requestOptions{})
}
