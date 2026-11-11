package wazobiamail

// Items is a collection envelope used by endpoints that return {"items":[...]}.
type Items[T any] struct {
	Items []T `json:"items"`
}

// Pagination is the API pagination metadata.
type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"pageSize"`
	TotalItems int  `json:"totalItems"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

// Page is a paginated collection envelope.
type Page[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

// SuccessResponse is returned by endpoints that only acknowledge an action.
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// URLAttachment is a publicly reachable attachment.
type URLAttachment struct {
	URL      string `json:"url"`
	FileName string `json:"fileName"`
}

// SendMessageRequest is the ergonomic send payload. The SDK converts it to the
// canonical nested content contract.
type SendMessageRequest struct {
	To             []string        `json:"-"`
	Cc             []string        `json:"-"`
	Bcc            []string        `json:"-"`
	From           string          `json:"-"`
	Subject        string          `json:"-"`
	HTML           string          `json:"-"`
	Text           string          `json:"-"`
	ReplyTo        string          `json:"-"`
	EnvelopeSender string          `json:"-"`
	Attachments    []URLAttachment `json:"-"`
	IdempotencyKey string          `json:"-"`
}

// SendResult is the receipt returned by the send endpoint.
type SendResult struct {
	Status     string `json:"status"`
	TrackingID string `json:"trackingId"`
	Recipients int    `json:"recipients"`
}

// SendEnvelope is the data envelope for a send response: {"result": {...}}.
type SendEnvelope struct {
	Result SendResult `json:"result"`
}

type sendContent struct {
	From     string `json:"from,omitempty"`
	Subject  string `json:"subject"`
	TextBody string `json:"textBody,omitempty"`
	HTMLBody string `json:"htmlBody,omitempty"`
	ReplyTo  string `json:"replyTo,omitempty"`
}

type sendPayload struct {
	To             []string        `json:"to"`
	Cc             []string        `json:"cc,omitempty"`
	Bcc            []string        `json:"bcc,omitempty"`
	Content        sendContent     `json:"content"`
	EnvelopeSender string          `json:"envelopeSender,omitempty"`
	Attachments    []URLAttachment `json:"attachments,omitempty"`
}

// SpamCheckRequest is a pre-flight spam/deliverability analysis request.
type SpamCheckRequest struct {
	Subject string            `json:"subject"`
	HTML    string            `json:"html,omitempty"`
	Text    string            `json:"text,omitempty"`
	From    string            `json:"from,omitempty"`
	ReplyTo string            `json:"replyTo,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// SpamRule is a single finding from the spam scanner.
type SpamRule struct {
	Code           string  `json:"code"`
	Points         float64 `json:"points"`
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	Recommendation string  `json:"recommendation"`
}

// SpamCheckResult is a spam/deliverability report.
type SpamCheckResult struct {
	Score               float64    `json:"score"`
	Threshold           float64    `json:"threshold"`
	IsSpam              bool       `json:"isSpam"`
	Rating              string     `json:"rating"`
	BadgeColor          string     `json:"badgeColor"`
	DeliverabilityScore int        `json:"deliverabilityScore"`
	Summary             string     `json:"summary"`
	Rules               []SpamRule `json:"rules"`
}

// TrackingMessage is a tracked outbound message.
type TrackingMessage struct {
	TrackingID          string  `json:"trackingId"`
	SenderDomain        string  `json:"senderDomain"`
	Subject             *string `json:"subject"`
	RecipientsCount     int     `json:"recipientsCount"`
	AcceptedCount       int     `json:"acceptedCount"`
	DeliveredCount      int     `json:"deliveredCount"`
	BouncedCount        int     `json:"bouncedCount"`
	OpenCount           int     `json:"openCount"`
	ClickCount          int     `json:"clickCount"`
	Status              string  `json:"status"`
	CreatedAt           string  `json:"createdAt"`
	LastOpenedAt        *string `json:"lastOpenedAt"`
	LastClickedAt       *string `json:"lastClickedAt"`
	LastDeliveryEventAt *string `json:"lastDeliveryEventAt"`
}

// TrackingSummary aggregates tracked message activity.
type TrackingSummary struct {
	Messages   int64 `json:"messages"`
	Recipients int64 `json:"recipients"`
	Accepted   int64 `json:"accepted"`
	Delivered  int64 `json:"delivered"`
	Bounced    int64 `json:"bounced"`
	Opens      int64 `json:"opens"`
	Clicks     int64 `json:"clicks"`
}

// TrackingList is the tracked messages response: {"messages":[],"summary":{}}.
type TrackingList struct {
	Messages []TrackingMessage `json:"messages"`
	Summary  TrackingSummary   `json:"summary"`
}

// TrackingEvent is a delivery/open/click event.
type TrackingEvent struct {
	ID             string         `json:"id"`
	EventType      string         `json:"eventType"`
	RecipientEmail *string        `json:"recipientEmail"`
	ClickURL       *string        `json:"clickUrl"`
	OccurredAt     string         `json:"occurredAt"`
	Engagement     map[string]any `json:"engagement"`
}

// TrackingDetail is a tracked message plus its event timeline.
type TrackingDetail struct {
	Message TrackingMessage `json:"message"`
	Events  []TrackingEvent `json:"events"`
}

// TrackingListParams filters tracked messages.
type TrackingListParams struct {
	Limit  int
	Offset int
	Status string
}

// EmailLog is a customer-visible delivery log row.
type EmailLog struct {
	ID         string  `json:"id"`
	TrackingID string  `json:"trackingId"`
	Recipient  *string `json:"recipient"`
	Subject    *string `json:"subject"`
	Status     string  `json:"status"`
	CreatedAt  string  `json:"createdAt"`
}

// EmailLogs is the email logs response.
type EmailLogs struct {
	Available             bool       `json:"available"`
	Messages              []EmailLog `json:"messages"`
	HasMore               bool       `json:"hasMore"`
	EventDetailsAvailable bool       `json:"eventDetailsAvailable"`
}

// EmailLogParams filters account email logs.
type EmailLogParams struct {
	Limit  int
	Offset int
	Status string
	From   string
	To     string
}
