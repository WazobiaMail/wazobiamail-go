package wazobiamail

import (
	"context"
	"net/http"
	"net/url"
)

// Domain is a sending domain.
type Domain struct {
	ID                   string  `json:"id"`
	AccountID            int64   `json:"accountId"`
	Domain               string  `json:"domain"`
	RequestedDomain      *string `json:"requestedDomain"`
	VerificationStatus   string  `json:"verificationStatus"`
	IsDefault            bool    `json:"isDefault"`
	DomainRole           string  `json:"domainRole"`
	ParentDomainID       *string `json:"parentDomainId"`
	CreateDefaultInboxes bool    `json:"createDefaultInboxes"`
	MailMode             string  `json:"mailMode"`
	CreatedAt            string  `json:"createdAt"`
}

// DomainEnvelope is {"domain":{...}}.
type DomainEnvelope struct {
	Domain Domain `json:"domain"`
}

// CreateDomainRequest describes a new sending domain.
type CreateDomainRequest struct {
	Domain                  string `json:"domain"`
	IsDefault               *bool  `json:"isDefault,omitempty"`
	UseRootDomain           *bool  `json:"useRootDomain,omitempty"`
	HostInboxesOnMainDomain *bool  `json:"hostInboxesOnMainDomain,omitempty"`
	SendingSubdomain        string `json:"sendingSubdomain,omitempty"`
	MailMode                string `json:"mailMode,omitempty"`
}

// UpdateDomainRequest updates a sending domain.
type UpdateDomainRequest struct {
	Domain        string `json:"domain,omitempty"`
	IsDefault     *bool  `json:"isDefault,omitempty"`
	UseRootDomain *bool  `json:"useRootDomain,omitempty"`
	MailMode      string `json:"mailMode,omitempty"`
}

// DNSRecord is one required DNS record.
type DNSRecord struct {
	ID            string  `json:"id"`
	RecordType    string  `json:"recordType"`
	Host          string  `json:"host"`
	Value         string  `json:"value"`
	TTL           int     `json:"ttl"`
	Status        string  `json:"status"`
	LastCheckedAt *string `json:"lastCheckedAt"`
}

// DomainDNS is the DNS payload returned by /dns-records.
type DomainDNS struct {
	Provisioning       bool        `json:"provisioning"`
	DomainID           string      `json:"domainId"`
	Domain             string      `json:"domain"`
	VerificationStatus string      `json:"verificationStatus"`
	Records            []DNSRecord `json:"records"`
	CopyBlock          string      `json:"copyBlock"`
}

// DomainDNSEnvelope is {"domain":{...}} for DNS records.
type DomainDNSEnvelope struct {
	Domain DomainDNS `json:"domain"`
}

// VerificationCheck is one checked DNS record.
type VerificationCheck struct {
	RecordType     string   `json:"recordType"`
	Host           string   `json:"host"`
	ExpectedValue  string   `json:"expectedValue"`
	Matched        bool     `json:"matched"`
	ObservedValues []string `json:"observedValues"`
}

// DomainVerification is the /verify result.
type DomainVerification struct {
	DomainID           string              `json:"domainId"`
	Domain             string              `json:"domain"`
	VerificationStatus string              `json:"verificationStatus"`
	VerifiedCount      int                 `json:"verifiedCount"`
	PendingCount       int                 `json:"pendingCount"`
	Checks             []VerificationCheck `json:"checks"`
}

// VerificationEnvelope is {"verification":{...}}.
type VerificationEnvelope struct {
	Verification DomainVerification `json:"verification"`
}

// CloudflareStatus describes the Cloudflare DNS connection.
type CloudflareStatus struct {
	Connected bool    `json:"connected"`
	ZoneID    *string `json:"zoneId"`
	Domain    *string `json:"domain"`
}

// CloudflareConnect is {"authorizationUrl":"..."}.
type CloudflareConnect struct {
	AuthorizationURL string `json:"authorizationUrl"`
}

// CloudflareSyncRequest asks Cloudflare to publish DNS records.
type CloudflareSyncRequest struct {
	DomainID string `json:"domainId"`
}

// CloudflareSyncResult is the sync response.
type CloudflareSyncResult struct {
	Zone               string `json:"zone"`
	Added              int    `json:"added"`
	Updated            int    `json:"updated"`
	Records            int    `json:"records"`
	VerificationStatus string `json:"verificationStatus"`
}

// DomainsService manages sending domains.
type DomainsService struct{ client *Client }

func (s *DomainsService) List(ctx context.Context) (*APIResponse[Items[Domain]], error) {
	return doJSON[Items[Domain]](ctx, s.client, http.MethodGet, "/sending/domains", nil, requestOptions{})
}

func (s *DomainsService) Create(ctx context.Context, req CreateDomainRequest) (*APIResponse[DomainEnvelope], error) {
	return doJSON[DomainEnvelope](ctx, s.client, http.MethodPost, "/sending/domains", req, requestOptions{})
}

func (s *DomainsService) Get(ctx context.Context, domainID string) (*APIResponse[DomainEnvelope], error) {
	return doJSON[DomainEnvelope](ctx, s.client, http.MethodGet, "/sending/domains/"+url.PathEscape(domainID), nil, requestOptions{})
}

func (s *DomainsService) Update(ctx context.Context, domainID string, req UpdateDomainRequest) (*APIResponse[DomainEnvelope], error) {
	return doJSON[DomainEnvelope](ctx, s.client, http.MethodPatch, "/sending/domains/"+url.PathEscape(domainID), req, requestOptions{})
}

func (s *DomainsService) Delete(ctx context.Context, domainID string) (*APIResponse[map[string]string], error) {
	return doJSON[map[string]string](ctx, s.client, http.MethodDelete, "/sending/domains/"+url.PathEscape(domainID), nil, requestOptions{})
}

func (s *DomainsService) DNSRecords(ctx context.Context, domainID string) (*APIResponse[DomainDNSEnvelope], error) {
	return doJSON[DomainDNSEnvelope](ctx, s.client, http.MethodGet, "/sending/domains/"+url.PathEscape(domainID)+"/dns-records", nil, requestOptions{})
}

func (s *DomainsService) Verify(ctx context.Context, domainID string, force bool) (*APIResponse[VerificationEnvelope], error) {
	return doJSON[VerificationEnvelope](ctx, s.client, http.MethodPost, "/sending/domains/"+url.PathEscape(domainID)+"/verify", map[string]bool{"force": force}, requestOptions{})
}

func (s *DomainsService) RefreshDNS(ctx context.Context, domainID string) (*APIResponse[VerificationEnvelope], error) {
	return doJSON[VerificationEnvelope](ctx, s.client, http.MethodPost, "/sending/domains/"+url.PathEscape(domainID)+"/dns-refresh", nil, requestOptions{})
}

// Cloudflare reads the Cloudflare DNS connection status.
func (s *DomainsService) CloudflareStatus(ctx context.Context, domainID string) (*APIResponse[CloudflareStatus], error) {
	opts := requestOptions{}
	if domainID != "" {
		opts.query = url.Values{"domainId": []string{domainID}}
	}
	return doJSON[CloudflareStatus](ctx, s.client, http.MethodGet, "/sending/domains/cloudflare/status", nil, opts)
}

// CloudflareConnect starts the Cloudflare OAuth connection.
func (s *DomainsService) CloudflareConnect(ctx context.Context, domainID string) (*APIResponse[CloudflareConnect], error) {
	opts := requestOptions{}
	if domainID != "" {
		opts.query = url.Values{"domainId": []string{domainID}}
	}
	return doJSON[CloudflareConnect](ctx, s.client, http.MethodGet, "/sending/domains/cloudflare/connect", nil, opts)
}

// CloudflareSync publishes DNS records to the connected Cloudflare zone.
func (s *DomainsService) CloudflareSync(ctx context.Context, domainID string) (*APIResponse[CloudflareSyncResult], error) {
	return doJSON[CloudflareSyncResult](ctx, s.client, http.MethodPost, "/sending/domains/cloudflare/sync", CloudflareSyncRequest{DomainID: domainID}, requestOptions{})
}

// CloudflareDisconnect removes the Cloudflare DNS connection.
func (s *DomainsService) CloudflareDisconnect(ctx context.Context, domainID string) (*APIResponse[map[string]any], error) {
	opts := requestOptions{}
	if domainID != "" {
		opts.query = url.Values{"domainId": []string{domainID}}
	}
	return doJSON[map[string]any](ctx, s.client, http.MethodDelete, "/sending/domains/cloudflare/disconnect", nil, opts)
}
