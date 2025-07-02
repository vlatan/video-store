package models

// User struct to store in the USER info in session
// A simplified version of goth.User
type User struct {
	ID             int    `json:"id"`
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	AvatarURL      string `json:"avatar_url"`
	AnalyticsID    string `json:"analytics_id"`
	LocalAvatarURL string `json:"local_avatar_url"`
	AccessToken    string `json:"access_token"`
}

// Check if user is authenticated
func (u *User) IsAuthenticated() bool {
	return u != nil && u.UserID != ""
}
