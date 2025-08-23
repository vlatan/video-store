package models

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"factual-docs/internal/config"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/r2"
	"fmt"
	"io"
	"log"
	"net/http"
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
		// Check if default avatar
		if avatar == defaultAvatar {
			return avatar
		}

		// Quick file existence check
		destination := filepath.Join(config.DataVolume, u.AnalyticsID+".jpg")
		if _, err := os.Stat(destination); err == nil {
			return avatar
		}

		// File missing, clear stale cache
		rdb.Delete(ctx, redisKey)
	}

	// Attempt to download the avatar, set default avatar on fail
	etag, err := u.DownloadAvatar(config, r2s)
	if err != nil {
		rdb.Set(ctx, redisKey, defaultAvatar, 24*7*time.Hour)
		return defaultAvatar
	}

	// Save avatar URL to Redis and return
	avatar = "/static/images/avatars/" + u.AnalyticsID + ".jpg?v=" + etag
	rdb.Set(ctx, redisKey, avatar, 24*7*time.Hour)
	return avatar
}

// Download remote image (user avatar)
func (u *User) DownloadAvatar(config *config.Config, r2s r2.Service) (string, error) {
	// Set the anaylytics ID in case it's missing
	if u.AnalyticsID == "" {
		u.SetAnalyticsID()
	}

	// Get remote file
	response, err := http.Get(u.AvatarURL)
	if err != nil {
		return "", fmt.Errorf("can't read the remote avatar file: %w", err)
	}
	defer response.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf(
			"failed to download avatar from %s: received status code %d",
			u.AvatarURL,
			response.StatusCode,
		)
	}

	// Create a file for writing
	destination := filepath.Join(config.DataVolume, u.AnalyticsID+".jpg")
	file, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("couldn't create file '%s': %w", destination, err)
	}

	// Flag to track if the download was successful
	valid := false

	// Run this clean up function on exit
	defer func() {
		if err := file.Close(); err != nil { // Close the file
			log.Printf("Warning: failed to close file '%s': %v", destination, err)
		}
		if !valid { // Remove the file if not successfuly created
			if err := os.Remove(destination); err != nil {
				log.Printf("Failed to remove partially created file '%s': %v", destination, err)
			}
		}
	}()

	// Init a hasher
	hasher := md5.New()

	// Create a multiwriter to write to the hasher and to the file
	multiWriter := io.MultiWriter(hasher, file)

	// Stream the response body directly into the hasher and the file
	_, err = io.Copy(multiWriter, response.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't hash or write to file '%s': %w", destination, err)
	}

	// Get the final hash sum and convert to a hex string
	hashInBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)

	valid = true
	return hashString, nil
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
