package models

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec G501
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

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
	ID             int
	ProviderUserId string
	Email          string
	Name           string
	Provider       string
	AvatarURL      string
	AnalyticsID    string
	LocalAvatarURL string
	AccessToken    string
	RefreshToken   string
	Expiry         time.Time
	LastSeen       *time.Time
	CreatedAt      *time.Time
	*config.Config
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

// Set the user analytics ID
func (u *User) SetAnalyticsID() {
	analyticsID := u.ProviderUserId + u.Provider + u.Email
	hashBytes := sha256.Sum256([]byte(analyticsID))
	u.AnalyticsID = fmt.Sprintf("%x", hashBytes)[:32]
}

// SetAvatar sets user avatar path, either from Redis,
// or downloads remote avatar, converts it to JPEG,
// uploads it to R2 and caches the path to Redis.
// The function will return only breaking errors,
// in this case only if the context expired.
// Every other error will be logged.
func (u *User) GetAvatar(
	ctx context.Context,
	config *config.Config,
	rdb *rdb.Service,
	r2s r2.Service) (string, error) {

	// Set the anaylytics ID in case it's missing
	if u.AnalyticsID == "" {
		u.SetAnalyticsID()
	}

	// Get avatar URL from Redis
	avatarKey := avatarCachePrefix + u.AnalyticsID
	r2URL, err := rdb.Client.Get(ctx, avatarKey).Result()

	// Return early if context error
	if utils.IsContextErr(err) {
		return "", err
	}

	// Avatar not in cache at all
	if err != nil || r2URL == "" {
		slog.Error(
			"failed to get avatar from Redis cache",
			"avatar", r2URL,
			"error", err,
		)
		r2URL = defaultAvatarPath
	}

	// Check if TTL for the avatar expired
	ttlKey := avatarCacheTTL + u.AnalyticsID
	ttl, err := rdb.Client.Exists(ctx, ttlKey).Result()

	// Return early if context error
	if utils.IsContextErr(err) {
		return "", err
	}

	if err != nil {
		slog.Error(
			"failed to get avatar's TTL from Redis cache",
			"avatar", r2URL,
			"error", err,
		)
	}

	// Cache timer haven't expired, return r2URL
	if ttl > 0 {
		return r2URL, nil
	}

	// The timer has expired at this point.
	// Refresh the avatar - reupload to R2 if changed.
	r2URL, err = u.refreshAvatar(ctx, config, r2s)

	// Return early if context error
	if utils.IsContextErr(err) {
		return "", err
	}

	if err != nil || r2URL == "" {
		slog.Error(
			"failed to refresh the avatar",
			"avatar", r2URL,
			"error", err,
		)
		r2URL = defaultAvatarPath
	}

	// Set the avatar URL in cache with long ttl only if not default avatar
	if r2URL != defaultAvatarPath {
		if err := rdb.Client.Set(ctx, avatarKey, r2URL, 30*24*time.Hour).Err(); err != nil {
			slog.Error(
				"failed to save the avatar in Redis",
				"redisKey", avatarKey,
				"avatarURL", r2URL,
				"error", err,
			)
		}

		// Return early if context error
		if utils.IsContextErr(err) {
			return "", err
		}
	}

	// Reset the timer
	if err := rdb.Client.Set(ctx, ttlKey, "true", 24*time.Hour).Err(); err != nil {
		slog.Error(
			"failed to reset the avatar TTL in Redis",
			"redisKey", ttlKey,
			"error", err,
		)
	}

	// Return early if context error
	if utils.IsContextErr(err) {
		return "", err
	}

	return r2URL, nil
}

// downloadAvatar downloads avatar from a remote source
func (u *User) downloadAvatar(ctx context.Context) ([]byte, error) {

	// Create a request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.AvatarURL, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create GET request for avatar %q download; %w",
			u.AvatarURL, err,
		)
	}

	// Execute the request
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to execute GET request for avatar %q; %w",
			u.AvatarURL, err,
		)
	}
	defer resp.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"received status code %d for avatar %q",
			resp.StatusCode, u.AvatarURL,
		)
	}

	// Limit size to prevent abuse (5MB max)
	limitedReader := io.LimitReader(resp.Body, 5*1024*1024)
	return io.ReadAll(limitedReader)
}

// refreshAvatar reuploads the user avatar at R2 if changed
func (u *User) refreshAvatar(
	ctx context.Context,
	config *config.Config,
	r2s r2.Service) (string, error) {

	// Download the avatar from remote location
	data, err := u.downloadAvatar(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to download avatar; %w", err)
	}

	// Hash the source image.
	// We are not doing anything important here, md5  is fine.
	sourceHash := fmt.Sprintf("%x", md5.Sum(data)) // #nosec G401

	// Get R2 object head
	head, err := r2s.HeadObject(
		ctx,
		config.R2CdnBucketName,
		fmt.Sprintf(avatarR2Path, u.AnalyticsID),
	)

	// Form the avatar URL
	avatarURL := &url.URL{
		Scheme:   "https",
		Host:     config.R2CdnDomain,
		Path:     fmt.Sprintf(avatarR2Path, u.AnalyticsID),
		RawQuery: "v=" + url.QueryEscape(sourceHash),
	}

	avatar := avatarURL.String()

	// If object source unchanged avatar not changed, return it
	if err == nil && head.Metadata != nil {
		storedHash, exists := head.Metadata["source-hash"]
		if exists && storedHash == sourceHash {
			return avatar, nil
		}
	}

	// Decode the avatar
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf(
			"failed to decode the file for avatar %s; %w",
			u.AvatarURL, err,
		)
	}

	// Convert to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf(
			"failed to convert the avatar %s to JPEG; %w",
			u.AvatarURL, err,
		)
	}

	// Upload object to bucket
	err = r2s.PutObject(
		ctx,
		config.R2CdnBucketName,
		fmt.Sprintf(avatarR2Path, u.AnalyticsID),
		bytes.NewReader(buf.Bytes()),
		"image/jpeg",
		map[string]string{"source-hash": sourceHash},
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to upload the avatar %s to bucket: %w",
			u.AnalyticsID, err,
		)
	}

	// Return the avatar
	return avatar, nil
}

// Delete avatar from object storage if exists
func (u *User) DeleteAvatar(
	ctx context.Context,
	config *config.Config,
	rdb *rdb.Service,
	r2s r2.Service) error {

	errs := make([]error, 0, 3)

	// Attemp to delete the avatar image from R2
	objectKey := fmt.Sprintf(avatarR2Path, u.AnalyticsID)
	err := r2s.DeleteObject(ctx, config.R2CdnBucketName, objectKey)
	err = fmt.Errorf("failed to remove avatar %q from R2: %w", objectKey, err)
	errs = append(errs, err)

	// Delete user and admin avatar Redis cache values
	for _, key := range []string{
		avatarCacheTTL + u.AnalyticsID,
		avatarCachePrefix + u.AnalyticsID,
	} {
		err := rdb.Client.Del(ctx, key).Err()
		err = fmt.Errorf("failed to remove avatar %q from Redis: %w", key, err)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
