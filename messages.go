package wazobiamail

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
)

// FileUpload is a binary attachment for the multipart send endpoint.
type FileUpload struct {
	FileName    string
	ContentType string
	Content     []byte
}

// UploadedFile is the attachment upload response.
type UploadedFile struct {
	FileName    string  `json:"fileName"`
	ContentType *string `json:"contentType"`
	SizeBytes   int64   `json:"sizeBytes"`
	PublicURL   *string `json:"publicUrl"`
}

// UploadedFileEnvelope is {"file":{...}}.
type UploadedFileEnvelope struct {
	File UploadedFile `json:"file"`
}

// MessagesService sends transactional email and reads tracking/log data.
type MessagesService struct{ client *Client }

// Send sends an email and returns its tracking receipt.
func (s *MessagesService) Send(ctx context.Context, req SendMessageRequest) (*APIResponse[SendEnvelope], error) {
	payload := buildSendPayload(req)
	return doJSON[SendEnvelope](ctx, s.client, http.MethodPost, "/sending/messages", payload, requestOptions{idempotencyKey: req.IdempotencyKey})
}

// SendWithAttachments sends a message with binary file parts.
func (s *MessagesService) SendWithAttachments(ctx context.Context, req SendMessageRequest, files []FileUpload) (*APIResponse[SendEnvelope], error) {
	payload := buildSendPayload(req)
	messageJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	body, headers, err := s.client.doMultipart(ctx, "/sending/messages/with-attachments", func(w *multipart.Writer) error {
		if err := w.WriteField("message", string(messageJSON)); err != nil {
			return err
		}
		for _, file := range files {
			part, err := w.CreateFormFile("files", file.FileName)
			if err != nil {
				return err
			}
			if _, err := part.Write(file.Content); err != nil {
				return err
			}
		}
		return nil
	}, requestOptions{idempotencyKey: req.IdempotencyKey})
	if err != nil {
		return nil, err
	}
	_ = headers
	return decodeBody[SendEnvelope](body)
}

// UploadAttachment uploads a file and returns its public URL.
func (s *MessagesService) UploadAttachment(ctx context.Context, file FileUpload) (*APIResponse[UploadedFileEnvelope], error) {
	body, _, err := s.client.doMultipart(ctx, "/sending/attachments/upload", func(w *multipart.Writer) error {
		part, err := w.CreateFormFile("file", file.FileName)
		if err != nil {
			return err
		}
		_, err = part.Write(file.Content)
		return err
	}, requestOptions{})
	if err != nil {
		return nil, err
	}
	return decodeBody[UploadedFileEnvelope](body)
}

// SpamCheck runs a pre-flight spam and deliverability analysis.
func (s *MessagesService) SpamCheck(ctx context.Context, req SpamCheckRequest) (*APIResponse[SpamCheckResult], error) {
	return doJSON[SpamCheckResult](ctx, s.client, http.MethodPost, "/sending/spam-check", req, requestOptions{})
}

// ListTracked lists tracked outgoing messages.
func (s *MessagesService) ListTracked(ctx context.Context, params TrackingListParams) (*APIResponse[TrackingList], error) {
	query := url.Values{}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		query.Set("offset", strconv.Itoa(params.Offset))
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	return doJSON[TrackingList](ctx, s.client, http.MethodGet, "/sending/tracking/messages", nil, requestOptions{query: query})
}

// GetTracked reads a tracked message and its event timeline.
func (s *MessagesService) GetTracked(ctx context.Context, trackingID string) (*APIResponse[TrackingDetail], error) {
	return doJSON[TrackingDetail](ctx, s.client, http.MethodGet, "/sending/tracking/messages/"+url.PathEscape(trackingID), nil, requestOptions{})
}

// ListEmailLogs reads customer-visible delivery logs.
func (s *MessagesService) ListEmailLogs(ctx context.Context, params EmailLogParams) (*APIResponse[EmailLogs], error) {
	query := url.Values{}
	if params.Limit > 0 {
		query.Set("limit", strconv.Itoa(params.Limit))
	}
	if params.Offset > 0 {
		query.Set("offset", strconv.Itoa(params.Offset))
	}
	if params.Status != "" {
		query.Set("status", params.Status)
	}
	if params.From != "" {
		query.Set("from", params.From)
	}
	if params.To != "" {
		query.Set("to", params.To)
	}
	return doJSON[EmailLogs](ctx, s.client, http.MethodGet, "/account/email-logs", nil, requestOptions{query: query})
}

func buildSendPayload(req SendMessageRequest) sendPayload {
	content := sendContent{Subject: req.Subject}
	if req.From != "" {
		content.From = req.From
	}
	if req.Text != "" {
		content.TextBody = req.Text
	}
	if req.HTML != "" {
		content.HTMLBody = req.HTML
	}
	if req.ReplyTo != "" {
		content.ReplyTo = req.ReplyTo
	}
	return sendPayload{
		To:             req.To,
		Cc:             req.Cc,
		Bcc:            req.Bcc,
		Content:        content,
		EnvelopeSender: req.EnvelopeSender,
		Attachments:    req.Attachments,
	}
}
