package types

import (

	// #nosec G501

	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	_ "image/gif" // Register GIF decoder
	_ "image/png" // Register PNG decoder

	"github.com/vlatan/video-store/internal/config"
	_ "golang.org/x/image/webp" // Register WebP decoder
)

// ============================================================================= //

// User struct to store the user data
type User struct {
	ID             int        `json:"-"`
	Provider       string     `json:"-"`
	ProviderUserId string     `json:"-"`
	Email          string     `json:"-"`
	Name           string     `json:"name,omitempty"`
	PublicID       string     `json:"public_id,omitempty"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	LocalAvatarURL string     `json:"local_avatar_url,omitempty"`
	AccessToken    string     `json:"-"`
	RefreshToken   string     `json:"-"`
	Expiry         time.Time  `json:"-"`
	LastSeen       *time.Time `json:"last_seen,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (u User) MarshalBinary() (data []byte, err error) {
	return json.Marshal(u)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (u *User) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

// IsAuthenticated reports whether the user is a real, loaded user
func (u *User) IsAuthenticated() bool {
	return u != nil &&
		u.ID != 0 &&
		u.Provider != "" &&
		u.ProviderUserId != ""
}

// IsAdmin reports whether the user matches the configured admin
func (u *User) IsAdmin(cfg *config.Config) bool {
	return u.IsAuthenticated() &&
		u.Provider == cfg.AdminProvider &&
		u.ProviderUserId == cfg.AdminProviderUserId
}

// Make a user public ID
func (u *User) MakePublicID() (string, error) {
	if u.Provider == "" || u.ProviderUserId == "" {
		return "", errors.New(
			"cannot generate public ID: provider and provider user id are required",
		)
	}

	publicID := fmt.Sprintf("%s:%s:%s", u.Provider, u.ProviderUserId, u.Email)
	hashBytes := sha256.Sum256([]byte(publicID))
	return fmt.Sprintf("%x", hashBytes), nil
}

// ============================================================================= //

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
