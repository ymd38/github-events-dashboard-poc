package repository

import (
	"database/sql"
	"fmt"

	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/model"
)

// UserRepository handles database operations for users.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByGitHubID returns a user by their GitHub ID, or nil if not found.
func (r *UserRepository) FindByGitHubID(githubID int64) (*model.User, error) {
	query := "SELECT id, github_id, login, display_name, avatar_url, access_token, last_login, created_at, updated_at FROM users WHERE github_id = ?"
	var u model.User
	err := r.db.QueryRow(query, githubID).Scan(
		&u.ID, &u.GitHubID, &u.Login, &u.DisplayName, &u.AvatarURL,
		&u.AccessToken, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by github_id: %w", err)
	}
	return &u, nil
}

// CreateOrUpdate inserts a new user or updates an existing one based on github_id.
func (r *UserRepository) CreateOrUpdate(user *model.User) (*model.User, error) {
	query := `INSERT INTO users (github_id, login, display_name, avatar_url, access_token, last_login)
		VALUES (?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			login = VALUES(login),
			display_name = VALUES(display_name),
			avatar_url = VALUES(avatar_url),
			access_token = VALUES(access_token),
			last_login = NOW()`
	result, err := r.db.Exec(query,
		user.GitHubID, user.Login, user.DisplayName, user.AvatarURL, user.AccessToken,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create or update user: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}
	if id == 0 {
		return r.FindByGitHubID(user.GitHubID)
	}
	user.ID = id
	return user, nil
}

// FindByID returns a user by their internal ID, or nil if not found.
func (r *UserRepository) FindByID(id int64) (*model.User, error) {
	query := "SELECT id, github_id, login, display_name, avatar_url, access_token, last_login, created_at, updated_at FROM users WHERE id = ?"
	var u model.User
	err := r.db.QueryRow(query, id).Scan(
		&u.ID, &u.GitHubID, &u.Login, &u.DisplayName, &u.AvatarURL,
		&u.AccessToken, &u.LastLogin, &u.CreatedAt, &u.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return &u, nil
}
