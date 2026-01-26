package models

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec G501
	"crypto/sha256"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/drivers/rdb"

	"github.com/vlatan/video-store/internal/integrations/r2"
	"github.com/vlatan/video-store/internal/utils"

	_ "image/gif" // Register GIF decoder
	_ "image/png" // Register PNG decoder

	_ "golang.org/x/image/webp" // Register WebP decoder
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

const avatarCacheKey = "avatar:r2:%s"
const avatarPath = "avatars/%s.jpg"
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
	hashBytes := sha256.Sum256([]byte(analyticsID))
	u.AnalyticsID = fmt.Sprintf("%x", hashBytes)[:32]
}

// SetAvatar gets user avatar path, either from redis,
// or downloads and stores avatar path to redis.
// If the function returns an error the default avatar might be set.
func (u *User) SetAvatar(
	ctx context.Context,
	config *config.Config,
	rdb *rdb.Service,
	r2s r2.Service,
	ttl time.Duration) error {

	// Set the anaylytics ID in case it's missing
	if u.AnalyticsID == "" {
		u.SetAnalyticsID()
	}

	// Get avatar URL from Redis
	redisKey := fmt.Sprintf(avatarCacheKey, u.AnalyticsID)
	avatar, err := rdb.Client.Get(ctx, redisKey).Result()

	if err == nil {
		u.LocalAvatarURL = avatar
		return nil
	}

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	// Set a slice of errors to acumulate non-breaking errors
	var errs []error

	// Save the error if not Redis nil error
	if !errors.Is(err, redis.Nil) {
		errs = append(errs, err)
	}

	// Attempt to download the avatar
	avatar, err = u.DownloadAvatar(ctx, config, r2s)

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	// Save non-nil error
	if err != nil {
		errs = append(errs, err)
		avatar = defaultAvatar
	}

	// Set avatar
	u.LocalAvatarURL = avatar
	err = rdb.Client.Set(ctx, redisKey, avatar, ttl).Err()

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	// Save non-nil error
	if err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
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
		return "", fmt.Errorf(
			"couldn't create request for avatar %s download: %w",
			u.AvatarURL, err,
		)
	}

	// Execute the request
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(
			"failed to download avatar %s: %w",
			u.AvatarURL, err,
		)
	}
	defer resp.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"failed to download avatar %s, received status code %d",
			u.AvatarURL, resp.StatusCode,
		)
	}

	// Read the body
	fileData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf(
			"failed to read file data for avatar %s: %w",
			u.AvatarURL, err,
		)
	}

	// Decode the avatar
	img, _, err := image.Decode(bytes.NewReader(fileData))
	if err != nil {
		return "", fmt.Errorf(
			"failed to decode the file for avatar %s: %w",
			u.AvatarURL, err,
		)
	}

	// Convert to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf(
			"failed to convert the avatar %s to JPEG: %w",
			u.AvatarURL, err,
		)
	}

	// Upload object to bucket
	err = r2s.PutObject(
		ctx,
		bytes.NewReader(buf.Bytes()),
		"image/jpeg",
		config.R2CdnBucketName,
		fmt.Sprintf(avatarPath, u.AnalyticsID),
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to upload the avatar %s to bucket: %w",
			u.AnalyticsID, err,
		)
	}

	etag := fmt.Sprintf("%x", md5.Sum(fileData)) // #nosec G401
	avatarURL := &url.URL{
		Scheme:   "https",
		Host:     config.R2CdnDomain,
		Path:     fmt.Sprintf(avatarPath, u.AnalyticsID),
		RawQuery: "v=" + url.QueryEscape(etag),
	}

	return avatarURL.String(), nil
}

// Delete local avatar if exists
func (u *User) DeleteAvatar(ctx context.Context, config *config.Config, rdb *rdb.Service, r2s r2.Service) {

	// Attemp to delete the avatar image from R2
	objectKey := fmt.Sprintf(avatarPath, u.AnalyticsID)
	if err := r2s.DeleteObject(ctx, config.R2CdnBucketName, objectKey); err != nil {
		log.Printf("Could not remove the avatar %s from R2: %v", objectKey, err)
	}

	redisKey := fmt.Sprintf(avatarCacheKey, u.AnalyticsID)
	if err := rdb.Client.Del(ctx, redisKey).Err(); err != nil {
		log.Printf("Could not remove the avatar %s from Redis: %v", redisKey, err)
	}
}
