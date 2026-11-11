package wazobiamail

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// Template is an email/WhatsApp template.
type Template struct {
	ID               string  `json:"id"`
	AccountID        *int64  `json:"accountId"`
	Name             string  `json:"name"`
	Slug             string  `json:"slug"`
	Channel          string  `json:"channel"`
	Category         string  `json:"category"`
	Status           string  `json:"status"`
	RejectionReason  *string `json:"rejectionReason"`
	Subject          string  `json:"subject"`
	Body             string  `json:"body"`
	TextBody         string  `json:"textBody"`
	MetaTemplateID   *string `json:"metaTemplateId"`
	MetaTemplateName *string `json:"metaTemplateName"`
	Language         string  `json:"language"`
	Components       any     `json:"components"`
	Variables        any     `json:"variables"`
	Design           any     `json:"design"`
	PreviewImageURL  *string `json:"previewImageUrl"`
	IsSystem         bool    `json:"isSystem"`
	Revision         int     `json:"revision"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
}

// TemplateList is the template page.
type TemplateList struct {
	Items  []Template `json:"items"`
	Total  int64      `json:"total"`
	Limit  int        `json:"limit"`
	Offset int        `json:"offset"`
}

// CreateTemplateRequest creates a template.
type CreateTemplateRequest struct {
	Name           string `json:"name"`
	Channel        string `json:"channel,omitempty"`
	Category       string `json:"category,omitempty"`
	Subject        string `json:"subject,omitempty"`
	Body           string `json:"body"`
	TextBody       string `json:"textBody,omitempty"`
	Language       string `json:"language,omitempty"`
	Components     any    `json:"components,omitempty"`
	Variables      any    `json:"variables,omitempty"`
	Design         any    `json:"design,omitempty"`
	AutoSubmitMeta bool   `json:"autoSubmitMeta,omitempty"`
	IsSystem       bool   `json:"isSystem,omitempty"`
}

// UpdateTemplateRequest updates a template.
type UpdateTemplateRequest struct {
	Name       string `json:"name,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Body       string `json:"body,omitempty"`
	TextBody   string `json:"textBody,omitempty"`
	Category   string `json:"category,omitempty"`
	Components any    `json:"components,omitempty"`
	Variables  any    `json:"variables,omitempty"`
	Design     any    `json:"design,omitempty"`
}

// PreviewTemplateRequest renders with sample variables.
type PreviewTemplateRequest struct {
	SampleVariables map[string]any `json:"sampleVariables,omitempty"`
}

// TemplatePreview is a rendered preview.
type TemplatePreview struct {
	Channel      string `json:"channel"`
	Subject      string `json:"subject"`
	RenderedBody string `json:"renderedBody"`
	RenderedText string `json:"renderedText"`
	Status       string `json:"status"`
	CanSend      bool   `json:"canSend"`
}

// TemplateListParams filters templates.
type TemplateListParams struct {
	Channel  string
	Category string
	Status   string
	Search   string
	Limit    int
	Offset   int
}

// TemplatesService manages templates (Enterprise/admin access).
type TemplatesService struct{ client *Client }

func (s *TemplatesService) List(ctx context.Context, params TemplateListParams) (*APIResponse[TemplateList], error) {
	query := url.Values{}
	if params.Channel != "" {
		query.Set("channel", params.Channel)
	}
	if params.Category != "" {
		query.Set("category", params.Category)
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.Search != "" {
		query.Set("search", params.Search)
	}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		query.Set("offset", strconv.Itoa(params.Offset))
	}
	return doJSON[TemplateList](ctx, s.client, http.MethodGet, "/templates", nil, requestOptions{query: query})
}

func (s *TemplatesService) Create(ctx context.Context, req CreateTemplateRequest) (*APIResponse[Template], error) {
	return doJSON[Template](ctx, s.client, http.MethodPost, "/templates", req, requestOptions{})
}

func (s *TemplatesService) Get(ctx context.Context, templateID string) (*APIResponse[Template], error) {
	return doJSON[Template](ctx, s.client, http.MethodGet, "/templates/"+url.PathEscape(templateID), nil, requestOptions{})
}

func (s *TemplatesService) Update(ctx context.Context, templateID string, req UpdateTemplateRequest) (*APIResponse[Template], error) {
	return doJSON[Template](ctx, s.client, http.MethodPut, "/templates/"+url.PathEscape(templateID), req, requestOptions{})
}

func (s *TemplatesService) Delete(ctx context.Context, templateID string) (*APIResponse[map[string]string], error) {
	return doJSON[map[string]string](ctx, s.client, http.MethodDelete, "/templates/"+url.PathEscape(templateID), nil, requestOptions{})
}

func (s *TemplatesService) Preview(ctx context.Context, templateID string, req PreviewTemplateRequest) (*APIResponse[TemplatePreview], error) {
	return doJSON[TemplatePreview](ctx, s.client, http.MethodPost, "/templates/"+url.PathEscape(templateID)+"/preview", req, requestOptions{})
}

func (s *TemplatesService) SyncMeta(ctx context.Context, templateID string) (*APIResponse[Template], error) {
	return doJSON[Template](ctx, s.client, http.MethodPost, "/templates/"+url.PathEscape(templateID)+"/sync-meta", map[string]any{}, requestOptions{})
}
