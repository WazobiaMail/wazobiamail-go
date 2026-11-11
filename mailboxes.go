package wazobiamail

import (
	"context"
	"net/http"
	"net/url"
)

// Mailbox is a hosted mailbox.
type Mailbox struct {
	ID                  string  `json:"id"`
	DomainID            string  `json:"domainId"`
	Address             string  `json:"address"`
	LocalPart           string  `json:"localPart"`
	DisplayName         *string `json:"displayName"`
	Status              string  `json:"status"`
	IsPlatformSubdomain bool    `json:"isPlatformSubdomain"`
	AutoProvisioned     bool    `json:"autoProvisioned"`
	Role                string  `json:"role"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
}

// MailboxEnvelope is {"mailbox":{...}}.
type MailboxEnvelope struct {
	Mailbox Mailbox `json:"mailbox"`
}

// CreateMailboxRequest provisions a mailbox.
type CreateMailboxRequest struct {
	DomainID    string `json:"domainId"`
	LocalPart   string `json:"localPart"`
	DisplayName string `json:"displayName,omitempty"`
	Password    string `json:"password,omitempty"`
	QuotaMB     int    `json:"quotaMb,omitempty"`
	Role        string `json:"role,omitempty"`
	Status      string `json:"status,omitempty"`
}

// UpdateMailboxRequest updates a mailbox.
type UpdateMailboxRequest struct {
	DisplayName string `json:"displayName,omitempty"`
	Password    string `json:"password,omitempty"`
}

// MailboxCredential is an SMTP/IMAP credential.
type MailboxCredential struct {
	ID        string  `json:"id"`
	MailboxID string  `json:"mailboxId"`
	Username  string  `json:"username"`
	ExpiresAt *string `json:"expiresAt"`
	RevokedAt *string `json:"revokedAt"`
	CreatedAt string  `json:"createdAt"`
}

// CreatedCredential is the inner credential + one-time password.
type CreatedCredential struct {
	Credential MailboxCredential `json:"credential"`
	Password   string            `json:"password"`
}

// CredentialEnvelope is {"credential":{"credential":...,"password":"..."}}.
type CredentialEnvelope struct {
	Credential CreatedCredential `json:"credential"`
}

// MailboxConnection contains client settings.
type MailboxConnection struct {
	MailboxID      string         `json:"mailboxId"`
	DomainID       string         `json:"domainId"`
	Username       string         `json:"username"`
	Status         string         `json:"status"`
	SMTP           map[string]any `json:"smtp"`
	IMAP           map[string]any `json:"imap"`
	POP3           map[string]any `json:"pop3"`
	Authentication string         `json:"authentication"`
	PasswordHint   string         `json:"passwordHint"`
}

// EmailAddress is a name/address pair.
type EmailAddress struct {
	Name  *string `json:"name"`
	Email string  `json:"email"`
}

// InboxAttachment is an attachment on an inbox message.
type InboxAttachment struct {
	BlobID    string `json:"blobId"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	SizeBytes int64  `json:"sizeBytes"`
}

// InboxMessage is a headless inbox message summary.
type InboxMessage struct {
	ID             string         `json:"id"`
	ThreadID       *string        `json:"threadId"`
	Subject        *string        `json:"subject"`
	Preview        *string        `json:"preview"`
	From           []EmailAddress `json:"from"`
	To             []EmailAddress `json:"to"`
	ReceivedAt     *string        `json:"receivedAt"`
	IsUnread       bool           `json:"isUnread"`
	HasAttachments bool           `json:"hasAttachments"`
	SizeBytes      int64          `json:"sizeBytes"`
}

// InboxMessageDetail includes bodies and attachments.
type InboxMessageDetail struct {
	InboxMessage
	CC          []EmailAddress    `json:"cc"`
	BCC         []EmailAddress    `json:"bcc"`
	ReplyTo     []EmailAddress    `json:"replyTo"`
	HTMLBody    *string           `json:"htmlBody"`
	TextBody    *string           `json:"textBody"`
	Attachments []InboxAttachment `json:"attachments"`
}

// MailboxMessages is the inbox list response.
type MailboxMessages struct {
	Items []InboxMessage `json:"items"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// ListMessagesParams filters inbox messages.
type ListMessagesParams struct {
	Page       int
	Limit      int
	UnreadOnly bool
	Search     string
}

// MailboxesService manages hosted mailboxes and enterprise inboxes.
type MailboxesService struct{ client *Client }

func (s *MailboxesService) List(ctx context.Context) (*APIResponse[[]Mailbox], error) {
	return doJSON[[]Mailbox](ctx, s.client, http.MethodGet, "/mailboxes", nil, requestOptions{})
}

func (s *MailboxesService) Create(ctx context.Context, req CreateMailboxRequest) (*APIResponse[Mailbox], error) {
	return doJSON[Mailbox](ctx, s.client, http.MethodPost, "/mailboxes", req, requestOptions{})
}

func (s *MailboxesService) Get(ctx context.Context, mailboxID string) (*APIResponse[MailboxEnvelope], error) {
	return doJSON[MailboxEnvelope](ctx, s.client, http.MethodGet, "/mailboxes/"+url.PathEscape(mailboxID), nil, requestOptions{})
}

func (s *MailboxesService) Update(ctx context.Context, mailboxID string, req UpdateMailboxRequest) (*APIResponse[MailboxEnvelope], error) {
	return doJSON[MailboxEnvelope](ctx, s.client, http.MethodPatch, "/mailboxes/"+url.PathEscape(mailboxID), req, requestOptions{})
}

func (s *MailboxesService) Delete(ctx context.Context, mailboxID string) (*APIResponse[map[string]string], error) {
	return doJSON[map[string]string](ctx, s.client, http.MethodDelete, "/mailboxes/"+url.PathEscape(mailboxID), nil, requestOptions{})
}

func (s *MailboxesService) Connection(ctx context.Context, mailboxID string) (*APIResponse[MailboxConnection], error) {
	return doJSON[MailboxConnection](ctx, s.client, http.MethodGet, "/mailboxes/"+url.PathEscape(mailboxID)+"/connection", nil, requestOptions{})
}

func (s *MailboxesService) CreateCredential(ctx context.Context, mailboxID string) (*APIResponse[CredentialEnvelope], error) {
	return doJSON[CredentialEnvelope](ctx, s.client, http.MethodPost, "/mailboxes/"+url.PathEscape(mailboxID)+"/credentials", map[string]any{}, requestOptions{})
}

func (s *MailboxesService) RotateCredential(ctx context.Context, mailboxID string) (*APIResponse[CredentialEnvelope], error) {
	return doJSON[CredentialEnvelope](ctx, s.client, http.MethodPost, "/mailboxes/"+url.PathEscape(mailboxID)+"/credentials/rotate", map[string]any{}, requestOptions{})
}

func (s *MailboxesService) RevokeCredential(ctx context.Context, mailboxID, credentialID string) (*APIResponse[map[string]string], error) {
	return doJSON[map[string]string](ctx, s.client, http.MethodPost, "/mailboxes/"+url.PathEscape(mailboxID)+"/credentials/"+url.PathEscape(credentialID)+"/revoke", nil, requestOptions{})
}

func (s *MailboxesService) Storage(ctx context.Context) (*APIResponse[map[string]any], error) {
	return doJSON[map[string]any](ctx, s.client, http.MethodGet, "/mailboxes/storage", nil, requestOptions{})
}

func (s *MailboxesService) ListMessages(ctx context.Context, mailboxID string, params ListMessagesParams) (*APIResponse[MailboxMessages], error) {
	query := url.Values{}
	if params.Page > 0 {
		query.Set("page", itoa(params.Page))
	}
	if params.Limit > 0 {
		query.Set("limit", itoa(params.Limit))
	}
	if params.UnreadOnly {
		query.Set("unreadOnly", "true")
	}
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	return doJSON[MailboxMessages](ctx, s.client, http.MethodGet, "/mailboxes/"+url.PathEscape(mailboxID)+"/messages", nil, requestOptions{query: query})
}

func (s *MailboxesService) GetMessage(ctx context.Context, mailboxID, messageID string) (*APIResponse[InboxMessageDetail], error) {
	return doJSON[InboxMessageDetail](ctx, s.client, http.MethodGet, "/mailboxes/"+url.PathEscape(mailboxID)+"/messages/"+url.PathEscape(messageID), nil, requestOptions{})
}

func (s *MailboxesService) DeleteMessage(ctx context.Context, mailboxID, messageID string) (*APIResponse[map[string]any], error) {
	return doJSON[map[string]any](ctx, s.client, http.MethodDelete, "/mailboxes/"+url.PathEscape(mailboxID)+"/messages/"+url.PathEscape(messageID), nil, requestOptions{})
}
