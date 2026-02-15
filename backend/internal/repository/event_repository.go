package repository

import (
	"database/sql"
	"fmt"

	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/model"
)

// EventRepository handles database operations for events.
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository creates a new EventRepository.
func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

// InsertEvent inserts a new event into the database.
// Returns the inserted event with its ID, or nil if it was a duplicate (idempotent).
func (r *EventRepository) InsertEvent(event *model.Event) (*model.Event, bool, error) {
	query := `INSERT INTO events (delivery_id, event_type, action, repo_name, sender_login, sender_avatar_url, title, body, html_url, event_data, occurred_at, received_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE id = LAST_INSERT_ID(id)`
	result, err := r.db.Exec(query,
		event.DeliveryID, event.EventType, event.Action, event.RepoName,
		event.SenderLogin, event.SenderAvatarURL, event.Title, event.Body,
		event.HTMLURL, event.EventData, event.OccurredAt, event.ReceivedAt,
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to insert event: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get last insert id: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("failed to get rows affected: %w", err)
	}
	event.ID = id
	isDuplicate := rowsAffected == 0
	return event, isDuplicate, nil
}

// ListEvents returns a paginated list of events, optionally filtered by event type.
func (r *EventRepository) ListEvents(page int, perPage int, eventType string) ([]model.Event, int, error) {
	countQuery := "SELECT COUNT(*) FROM events"
	listQuery := "SELECT id, delivery_id, event_type, action, repo_name, sender_login, sender_avatar_url, title, body, html_url, event_data, occurred_at, received_at, created_at FROM events"
	var args []interface{}
	if eventType != "" {
		countQuery += " WHERE event_type = ?"
		listQuery += " WHERE event_type = ?"
		args = append(args, eventType)
	}
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count events: %w", err)
	}
	offset := (page - 1) * perPage
	listQuery += " ORDER BY received_at DESC LIMIT ? OFFSET ?"
	listArgs := append(args, perPage, offset)
	rows, err := r.db.Query(listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list events: %w", err)
	}
	defer rows.Close()
	var events []model.Event
	for rows.Next() {
		var e model.Event
		if err := rows.Scan(
			&e.ID, &e.DeliveryID, &e.EventType, &e.Action, &e.RepoName,
			&e.SenderLogin, &e.SenderAvatarURL, &e.Title, &e.Body,
			&e.HTMLURL, &e.EventData, &e.OccurredAt, &e.ReceivedAt, &e.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, e)
	}
	if events == nil {
		events = []model.Event{}
	}
	return events, total, nil
}

// GetEventByID returns a single event by its ID.
func (r *EventRepository) GetEventByID(id int64) (*model.Event, error) {
	query := "SELECT id, delivery_id, event_type, action, repo_name, sender_login, sender_avatar_url, title, body, html_url, event_data, occurred_at, received_at, created_at FROM events WHERE id = ?"
	var e model.Event
	err := r.db.QueryRow(query, id).Scan(
		&e.ID, &e.DeliveryID, &e.EventType, &e.Action, &e.RepoName,
		&e.SenderLogin, &e.SenderAvatarURL, &e.Title, &e.Body,
		&e.HTMLURL, &e.EventData, &e.OccurredAt, &e.ReceivedAt, &e.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return &e, nil
}
