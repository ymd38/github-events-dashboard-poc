package model

import "time"

// User represents a dashboard user authenticated via GitHub OAuth.
type User struct {
	ID          int64     `json:"id"`
	GitHubID    int64     `json:"github_id"`
	Login       string    `json:"login"`
	DisplayName string    `json:"display_name"`
	AvatarURL   *string   `json:"avatar_url"`
	AccessToken string    `json:"-"`
	LastLogin   time.Time `json:"last_login"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserResponse represents user information returned by the API (no sensitive fields).
type UserResponse struct {
	ID          int64   `json:"id"`
	Login       string  `json:"login"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

// ToResponse converts a User to a UserResponse, stripping sensitive fields.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		Login:       u.Login,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}
}
