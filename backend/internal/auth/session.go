package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	sessionName    = "github-dashboard-session"
	sessionUserID  = "user_id"
	sessionMaxAge  = 86400 // 24 hours
)

// SessionManager manages user sessions using gorilla/sessions.
type SessionManager struct {
	store *sessions.CookieStore
}

// NewSessionManager creates a new SessionManager with the given secret.
func NewSessionManager(secret string) *SessionManager {
	store := sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   sessionMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return &SessionManager{store: store}
}

// SetUserID stores the user ID in the session.
func (sm *SessionManager) SetUserID(w http.ResponseWriter, r *http.Request, userID int64) error {
	session, err := sm.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Values[sessionUserID] = userID
	return session.Save(r, w)
}

// GetUserID retrieves the user ID from the session. Returns 0 if not found.
func (sm *SessionManager) GetUserID(r *http.Request) (int64, error) {
	session, err := sm.store.Get(r, sessionName)
	if err != nil {
		return 0, err
	}
	val, ok := session.Values[sessionUserID]
	if !ok {
		return 0, nil
	}
	userID, ok := val.(int64)
	if !ok {
		return 0, nil
	}
	return userID, nil
}

// Destroy removes the session.
func (sm *SessionManager) Destroy(w http.ResponseWriter, r *http.Request) error {
	session, err := sm.store.Get(r, sessionName)
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
