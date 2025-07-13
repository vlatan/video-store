package models

import "time"

// User struct to store in the USER info in session
// A simplified version of goth.User
type User struct {
	ID             int        `json:"id,omitempty"`
	UserID         string     `json:"user_id,omitempty"`
	Email          string     `json:"email,omitempty"`
	Name           string     `json:"name,omitempty"`
	Provider       string     `json:"provider"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	AnalyticsID    string     `json:"analytics_id,omitempty"`
	LocalAvatarURL string     `json:"local_avatar_url,omitempty"`
	AccessToken    string     `json:"access_token,omitempty"`
	LastSeen       *time.Time `json:"last_seen,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
}

// Check if user is authenticated
func (u *User) IsAuthenticated() bool {
	return u != nil && u.UserID != ""
}
