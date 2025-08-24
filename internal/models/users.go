package models

import (
	"bytes"
	"context"
	"crypto/md5"
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/r2"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// Collection of users
type Users struct {
	TotalNum int
	Items    []User
}

// User struct to store in the USER info in session
type User struct {
	ID             int        `json:"id,omitempty"`
	ProviderUserId string     `json:"user_id,omitempty"`
	Email          string     `json:"email,omitempty"`
	Name           string     `json:"name,omitempty"`
	Provider       string     `json:"provider"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	AnalyticsID    string     `json:"analytics_id,omitempty"`
	LocalAvatarURL string     `json:"local_avatar_url,omitempty"`
	AccessToken    string     `json:"access_token,omitempty"`
	RefreshToken   string     `json:"refresh_token,omitempty"`
	Expiry         time.Time  `json:"expiry"`
	LastSeen       *time.Time `json:"last_seen,omitempty"`
	CreatedAt      *time.Time `json:"created_at,omitempty"`
}

const avatarCacheKey = "avatar:%s"
const defaultAvatar = "/static/images/default-avatar.jpg"

var avatarTimeout time.Duration = 24 * time.Hour

// Check if the user is authenticated
func (u *User) IsAuthenticated() bool {
	return u != nil && u.ProviderUserId != ""
}

// Check if the user is Admin
func (u *User) IsAdmin(adminID, adminProvider string) bool {
	return u.IsAuthenticated() &&
		u.ProviderUserId == adminID &&
		u.Provider == adminProvider
}

// Set the user analytics ID
func (u *User) SetAnalyticsID() {
	analyticsID := u.ProviderUserId + u.Provider + u.Email
	u.AnalyticsID = fmt.Sprintf("%x", md5.Sum([]byte(analyticsID)))
}

// Get user avatar path, either from redis, or download and store avatar path to redis
func (u *User) GetAvatar(
	ctx context.Context,
	config *config.Config,
	rdb redis.Service,
	r2s r2.Service) string {

	// Set the anaylytics ID in case it's missing
	if u.AnalyticsID == "" {
		u.SetAnalyticsID()
	}

	// Get avatar URL from Redis
	redisKey := fmt.Sprintf(avatarCacheKey, u.AnalyticsID)
	avatar, err := rdb.Get(ctx, redisKey)
	if err == nil {
		return avatar
	}

	// Attempt to download the avatar, set default avatar on fail
	etag, err := u.DownloadAvatar(ctx, config, r2s)
	if err != nil {
		rdb.Set(ctx, redisKey, defaultAvatar, avatarTimeout)
		return defaultAvatar
	}

	avatarURL := &url.URL{
		Scheme:   "https",
		Host:     config.R2CdnDomain,
		Path:     fmt.Sprintf("/avatars/%s.jpg", u.AnalyticsID),
		RawQuery: "v=" + url.QueryEscape(etag),
	}

	// Save avatar URL to Redis and return
	rdb.Set(ctx, redisKey, avatarURL.String(), avatarTimeout)
	return avatar
}

// Download remote image (user avatar)
func (u *User) DownloadAvatar(ctx context.Context, config *config.Config, r2s r2.Service) (string, error) {
	// Set the anaylytics ID in case it's missing
	if u.AnalyticsID == "" {
		u.SetAnalyticsID()
	}

	// Create a request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.AvatarURL, nil)
	if err != nil {
		return "", fmt.Errorf("couldn't create request for avatar download: %w", err)
	}

	// Execute the request
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download avatar: %w", err)
	}
	defer resp.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed to download avatar from, received status code %d",
			resp.StatusCode,
		)
	}

	// Read the body
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read file data: %w", err)
	}

	// Determine content type
	contentType := http.DetectContentType(fileData)

	// Upload object to bucket
	err = r2s.PutObject(
		ctx,
		bytes.NewReader(fileData),
		contentType,
		config.R2CdnBucketName,
		fmt.Sprintf("/avatars/%s.jpg", u.AnalyticsID),
	)

	if err != nil {
		return "", fmt.Errorf("failed to upload the object: %w", err)
	}

	etag := fmt.Sprintf("%x", md5.Sum(fileData))
	return etag, nil
}

// Delete local avatar if exists
func (u *User) DeleteAvatar(ctx context.Context, rdb redis.Service, config *config.Config) {
	avatarPath := filepath.Join(config.DataVolume, u.AnalyticsID+".jpg")
	if err := os.Remove(avatarPath); err != nil && err != os.ErrNotExist {
		log.Printf("Could not remove the local avatar %s: %v", avatarPath, err)
	}

	redisKey := fmt.Sprintf(avatarCacheKey, u.AnalyticsID)
	if err := rdb.Delete(ctx, redisKey); err != nil {
		log.Printf("Could not remove the avatar %s from Redis: %v", redisKey, err)
	}
}
