package model

import "time"

// Event represents a GitHub webhook event stored in the database.
type Event struct {
	ID              int64      `json:"id"`
	DeliveryID      string     `json:"delivery_id"`
	EventType       string     `json:"event_type"`
	Action          string     `json:"action"`
	RepoName        string     `json:"repo_name"`
	SenderLogin     string     `json:"sender_login"`
	SenderAvatarURL *string    `json:"sender_avatar_url"`
	Title           *string    `json:"title"`
	Body            *string    `json:"body"`
	HTMLURL         string     `json:"html_url"`
	EventData       *string    `json:"event_data"`
	OccurredAt      time.Time  `json:"occurred_at"`
	ReceivedAt      time.Time  `json:"received_at"`
	CreatedAt       time.Time  `json:"created_at"`
}

// EventListResponse represents a paginated list of events returned by the API.
type EventListResponse struct {
	Events     []Event    `json:"events"`
	Pagination Pagination `json:"pagination"`
}

// Pagination holds pagination metadata.
type Pagination struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// WebhookResponse represents the response returned after processing a webhook.
type WebhookResponse struct {
	Status  string `json:"status"`
	EventID *int64 `json:"event_id,omitempty"`
}
