package wazobiamail

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// SupportAttachment is a ticket attachment reference.
type SupportAttachment struct {
	ID          string `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	SizeBytes   int64  `json:"sizeBytes"`
	URL         string `json:"url"`
}

// SupportUpload is an inline base64 attachment (max 5 MiB per file).
type SupportUpload struct {
	Filename string `json:"filename"`
	Data     string `json:"data"`
}

// SupportMessage is one message in a ticket thread.
type SupportMessage struct {
	ID          string              `json:"id"`
	TicketID    string              `json:"ticketId"`
	SenderName  string              `json:"senderName"`
	SenderEmail string              `json:"senderEmail"`
	Message     string              `json:"message"`
	IsFromStaff bool                `json:"isFromStaff"`
	ContentType string              `json:"contentType"`
	Attachments []SupportAttachment `json:"attachments"`
	CreatedAt   string              `json:"createdAt"`
}

// SupportTicket is a support ticket with its thread.
type SupportTicket struct {
	ID             string           `json:"id"`
	AccountID      *int64           `json:"accountId"`
	Name           string           `json:"name"`
	Email          string           `json:"email"`
	Subject        string           `json:"subject"`
	Domain         *string          `json:"domain"`
	Category       string           `json:"category"`
	Priority       string           `json:"priority"`
	Status         string           `json:"status"`
	Messages       []SupportMessage `json:"messages"`
	CustomerReadAt *string          `json:"customerReadAt"`
	StaffReadAt    *string          `json:"staffReadAt"`
	CreatedAt      string           `json:"createdAt"`
	UpdatedAt      string           `json:"updatedAt"`
}

// SupportTicketList is the paginated ticket list.
type SupportTicketList struct {
	Items []SupportTicket `json:"items"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Pages int             `json:"pages"`
}

// CreateSupportTicketRequest opens a ticket.
type CreateSupportTicketRequest struct {
	Name        string          `json:"name"`
	Email       string          `json:"email"`
	Subject     string          `json:"subject"`
	Domain      string          `json:"domain,omitempty"`
	Category    string          `json:"category"`
	Priority    string          `json:"priority"`
	Message     string          `json:"message"`
	ContentType string          `json:"contentType,omitempty"`
	Attachments []SupportUpload `json:"attachments,omitempty"`
}

// CreatedSupportTicket is the ticket creation receipt.
type CreatedSupportTicket struct {
	ID       string `json:"id"`
	TicketID string `json:"ticketId"`
}

// ReplySupportTicketRequest replies to a ticket.
type ReplySupportTicketRequest struct {
	Message     string          `json:"message"`
	ContentType string          `json:"contentType,omitempty"`
	Attachments []SupportUpload `json:"attachments,omitempty"`
}

// SupportListParams filters tickets.
type SupportListParams struct {
	Status     string
	Page       int
	Limit      int
	Sort       string
	Descending bool
}

// SupportService manages support tickets.
type SupportService struct{ client *Client }

func (s *SupportService) Create(ctx context.Context, req CreateSupportTicketRequest) (*APIResponse[CreatedSupportTicket], error) {
	return doJSON[CreatedSupportTicket](ctx, s.client, http.MethodPost, "/support", req, requestOptions{})
}

func (s *SupportService) List(ctx context.Context, params SupportListParams) (*APIResponse[SupportTicketList], error) {
	query := url.Values{}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Page > 0 {
		query.Set("page", strconv.Itoa(params.Page))
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Sort != "" {
		query.Set("sort", params.Sort)
	}
	if params.Descending {
		query.Set("descending", "true")
	}
	return doJSON[SupportTicketList](ctx, s.client, http.MethodGet, "/support/my", nil, requestOptions{query: query})
}

func (s *SupportService) Get(ctx context.Context, ticketID string) (*APIResponse[SupportTicket], error) {
	return doJSON[SupportTicket](ctx, s.client, http.MethodGet, "/support/tickets/"+url.PathEscape(ticketID), nil, requestOptions{})
}

func (s *SupportService) Reply(ctx context.Context, ticketID string, req ReplySupportTicketRequest) (*APIResponse[SupportMessage], error) {
	return doJSON[SupportMessage](ctx, s.client, http.MethodPost, "/support/tickets/"+url.PathEscape(ticketID)+"/messages", req, requestOptions{})
}

func (s *SupportService) MarkRead(ctx context.Context, ticketID string) (*APIResponse[map[string]any], error) {
	return doJSON[map[string]any](ctx, s.client, http.MethodPost, "/support/tickets/"+url.PathEscape(ticketID)+"/read", map[string]any{}, requestOptions{})
}

func (s *SupportService) DownloadAttachment(ctx context.Context, attachmentID string) ([]byte, http.Header, error) {
	return s.client.doBytes(ctx, "/support/attachments/"+url.PathEscape(attachmentID), requestOptions{})
}
