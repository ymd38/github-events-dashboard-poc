package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/middleware"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/model"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/repository"
)

const maxBodyLength = 500

// EventService handles business logic for webhook event processing.
type EventService struct {
	repo *repository.EventRepository
}

// NewEventService creates a new EventService.
func NewEventService(repo *repository.EventRepository) *EventService {
	return &EventService{repo: repo}
}

// ProcessWebhook parses a GitHub webhook payload and stores the event.
// Returns the response, the saved event (nil if duplicate/ignored), and an error.
func (s *EventService) ProcessWebhook(deliveryID string, eventType string, payload []byte) (*model.WebhookResponse, *model.Event, error) {
	event, err := s.parsePayload(deliveryID, eventType, payload)
	if err != nil {
		return nil, nil, err
	}
	if event == nil {
		return &model.WebhookResponse{Status: "ignored"}, nil, nil
	}
	saved, isDuplicate, err := s.repo.InsertEvent(event)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to save event: %w", err)
	}
	if isDuplicate {
		middleware.LogEvent("info", "duplicate event ignored", map[string]interface{}{
			"delivery_id": deliveryID,
		})
		return &model.WebhookResponse{Status: "duplicate"}, nil, nil
	}
	middleware.LogEvent("info", "event saved", map[string]interface{}{
		"delivery_id": deliveryID,
		"event_id":    saved.ID,
		"event_type":  saved.EventType,
	})
	return &model.WebhookResponse{Status: "received", EventID: &saved.ID}, saved, nil
}

// ListEvents returns a paginated list of events.
func (s *EventService) ListEvents(page int, perPage int, eventType string) (*model.EventListResponse, error) {
	events, total, err := s.repo.ListEvents(page, perPage, eventType)
	if err != nil {
		return nil, err
	}
	totalPages := (total + perPage - 1) / perPage
	return &model.EventListResponse{
		Events: events,
		Pagination: model.Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	}, nil
}

// GetEventByID returns a single event by ID.
func (s *EventService) GetEventByID(id int64) (*model.Event, error) {
	return s.repo.GetEventByID(id)
}

func (s *EventService) parsePayload(deliveryID string, eventType string, payload []byte) (*model.Event, error) {
	switch eventType {
	case "issues":
		return s.parseIssueEvent(deliveryID, payload)
	case "pull_request":
		return s.parsePullRequestEvent(deliveryID, payload)
	default:
		return nil, nil
	}
}

type issuePayload struct {
	Action string `json:"action"`
	Issue  struct {
		Title   string `json:"title"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
	} `json:"issue"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Sender struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"sender"`
}

func (s *EventService) parseIssueEvent(deliveryID string, payload []byte) (*model.Event, error) {
	var p issuePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to parse issue payload: %w", err)
	}
	if p.Action != "opened" {
		return nil, nil
	}
	body := truncateString(p.Issue.Body, maxBodyLength)
	return &model.Event{
		DeliveryID:      deliveryID,
		EventType:       "issues",
		Action:          p.Action,
		RepoName:        p.Repository.FullName,
		SenderLogin:     p.Sender.Login,
		SenderAvatarURL: ptrString(p.Sender.AvatarURL),
		Title:           ptrString(p.Issue.Title),
		Body:            ptrString(body),
		HTMLURL:         p.Issue.HTMLURL,
		OccurredAt:      time.Now().UTC(),
		ReceivedAt:      time.Now().UTC(),
	}, nil
}

type pullRequestPayload struct {
	Action      string `json:"action"`
	PullRequest struct {
		Title   string `json:"title"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		Merged  bool   `json:"merged"`
	} `json:"pull_request"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Sender struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"sender"`
}

func (s *EventService) parsePullRequestEvent(deliveryID string, payload []byte) (*model.Event, error) {
	var p pullRequestPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, fmt.Errorf("failed to parse pull_request payload: %w", err)
	}
	if p.Action != "closed" || !p.PullRequest.Merged {
		return nil, nil
	}
	body := truncateString(p.PullRequest.Body, maxBodyLength)
	return &model.Event{
		DeliveryID:      deliveryID,
		EventType:       "pull_request",
		Action:          "merged",
		RepoName:        p.Repository.FullName,
		SenderLogin:     p.Sender.Login,
		SenderAvatarURL: ptrString(p.Sender.AvatarURL),
		Title:           ptrString(p.PullRequest.Title),
		Body:            ptrString(body),
		HTMLURL:         p.PullRequest.HTMLURL,
		OccurredAt:      time.Now().UTC(),
		ReceivedAt:      time.Now().UTC(),
	}, nil
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
