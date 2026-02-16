package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/middleware"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/model"
	"github.com/hirokazuyamada/github-events-dashboard-poc/backend/internal/repository"
)

const stateLength = 16

// OAuthHandler handles GitHub OAuth authentication endpoints.
type OAuthHandler struct {
	oauthConfig    *oauth2.Config
	sessionManager *SessionManager
	userRepo       *repository.UserRepository
	frontendURL    string
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(clientID string, clientSecret string, frontendURL string, sessionManager *SessionManager, userRepo *repository.UserRepository) *OAuthHandler {
	return &OAuthHandler{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"read:user"},
			Endpoint:     github.Endpoint,
		},
		sessionManager: sessionManager,
		userRepo:       userRepo,
		frontendURL:    frontendURL,
	}
}

// Login handles GET /api/auth/login - redirects to GitHub authorization URL.
func (h *OAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		http.Error(w, "failed to generate state", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	url := h.oauthConfig.AuthCodeURL(state)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback handles GET /api/auth/callback - exchanges code for token, creates session.
func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		middleware.LogEvent("warn", "OAuth state mismatch", nil)
		http.Redirect(w, r, h.frontendURL+"/login?error=state_mismatch", http.StatusTemporaryRedirect)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, h.frontendURL+"/login?error=no_code", http.StatusTemporaryRedirect)
		return
	}
	token, err := h.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		middleware.LogEvent("error", "OAuth token exchange failed", map[string]interface{}{"error": err.Error()})
		http.Redirect(w, r, h.frontendURL+"/login?error=token_exchange", http.StatusTemporaryRedirect)
		return
	}
	ghUser, err := h.fetchGitHubUser(token.AccessToken)
	if err != nil {
		middleware.LogEvent("error", "failed to fetch GitHub user", map[string]interface{}{"error": err.Error()})
		http.Redirect(w, r, h.frontendURL+"/login?error=user_fetch", http.StatusTemporaryRedirect)
		return
	}
	user := &model.User{
		GitHubID:    ghUser.ID,
		Login:       ghUser.Login,
		DisplayName: ghUser.Name,
		AvatarURL:   ptrString(ghUser.AvatarURL),
		AccessToken: token.AccessToken,
	}
	saved, err := h.userRepo.CreateOrUpdate(user)
	if err != nil {
		middleware.LogEvent("error", "failed to save user", map[string]interface{}{"error": err.Error()})
		http.Redirect(w, r, h.frontendURL+"/login?error=save_user", http.StatusTemporaryRedirect)
		return
	}
	if err := h.sessionManager.SetUserID(w, r, saved.ID); err != nil {
		middleware.LogEvent("error", "failed to create session", map[string]interface{}{"error": err.Error()})
		http.Redirect(w, r, h.frontendURL+"/login?error=session", http.StatusTemporaryRedirect)
		return
	}
	middleware.LogEvent("info", "user logged in", map[string]interface{}{
		"user_id": saved.ID,
		"login":   saved.Login,
	})
	http.Redirect(w, r, h.frontendURL, http.StatusTemporaryRedirect)
}

// Logout handles POST /api/auth/logout - destroys the session.
func (h *OAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.sessionManager.Destroy(w, r); err != nil {
		middleware.LogEvent("error", "failed to destroy session", map[string]interface{}{"error": err.Error()})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
}

// Me handles GET /api/auth/me - returns current user info.
func (h *OAuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID, err := h.sessionManager.GetUserID(r)
	if err != nil || userID == 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
		return
	}
	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "user not found"})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.ToResponse())
}

type gitHubUserResponse struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

func (h *OAuthHandler) fetchGitHubUser(accessToken string) (*gitHubUserResponse, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	var ghUser gitHubUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		return nil, err
	}
	return &ghUser, nil
}

func generateState() (string, error) {
	b := make([]byte, stateLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
