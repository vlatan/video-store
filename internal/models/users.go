package models

import (

	// #nosec G501
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vlatan/video-store/internal/config"

	_ "image/gif" // Register GIF decoder
	_ "image/png" // Register PNG decoder

	_ "golang.org/x/image/webp" // Register WebP decoder
)

// We need an interface to avoid circular imports
type Queuer interface {
	Enqueue(u *User) bool
}

// Collection of users
type Users struct {
	TotalNum int
	Items    []User
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (u Users) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (u *Users) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

// User struct to store in the USER info in session
type User struct {
	ID             int            `json:"-"`
	ProviderUserId string         `json:"-"`
	Email          string         `json:"-"`
	Name           string         `json:"name,omitempty"`
	Provider       string         `json:"-"`
	AvatarURL      string         `json:"avatar_url,omitempty"`
	PublicID       string         `json:"public_id,omitempty"`
	LocalAvatarURL string         `json:"local_avatar_url,omitempty"`
	AccessToken    string         `json:"-"`
	RefreshToken   string         `json:"-"`
	Expiry         time.Time      `json:"-"`
	LastSeen       *time.Time     `json:"last_seen,omitempty"`
	CreatedAt      *time.Time     `json:"created_at,omitempty"`
	Config         *config.Config `json:"-"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (u User) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

// Check if the user is authenticated
func (u *User) IsAuthenticated() bool {
	return u != nil && u.ProviderUserId != ""
}

// Check if the user is Admin
func (u *User) IsAdmin() bool {
	return u.IsAuthenticated() &&
		u.ProviderUserId == u.Config.AdminProviderUserId &&
		u.Provider == u.Config.AdminProvider
}

// Set the user public ID
func (u *User) MakePublicID() string {
	publicID := u.ProviderUserId + u.Provider + u.Email
	hashBytes := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("%x", hashBytes)[:32]
}
