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

const AvatarAdminPrefix = "avatar:admin:"
const AvatarUserPrefix = "avatar:user:"
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
func (u *User) GetAvatar(
	ctx context.Context,
	config *config.Config,
	rdb *rdb.Service,
	r2s r2.Service,
	keyPrefix string,
	ttl time.Duration) error {

	// Set the anaylytics ID in case it's missing
	if u.AnalyticsID == "" {
		u.SetAnalyticsID()
	}

	// Get avatar URL from Redis
	redisKey := keyPrefix + u.AnalyticsID
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

	// Refresh the avatar
	avatar, err = u.refreshAvatar(ctx, config, rdb, r2s)

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

func (u *User) downloadAvatar(ctx context.Context) ([]byte, error) {

	// Create a request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.AvatarURL, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"couldn't create request for avatar %q download; %w",
			u.AvatarURL, err,
		)
	}

	// Execute the request
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to download avatar %q; %w",
			u.AvatarURL, err,
		)
	}
	defer resp.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to download avatar %s, received status code %d",
			u.AvatarURL, resp.StatusCode,
		)
	}

	// Limit size to prevent abuse (5MB max)
	limitedReader := io.LimitReader(resp.Body, 5*1024*1024)
	return io.ReadAll(limitedReader)
}

// updateCache sets both user and admin avatar caches
func (u *User) updateAvatarCache(ctx context.Context, rdb *rdb.Service, url string) (string, error) {

	userKey := AvatarUserPrefix + u.AnalyticsID
	adminKey := AvatarAdminPrefix + u.AnalyticsID

	// Set user cache (1 day)
	if err := rdb.Client.Set(ctx, userKey, url, 24*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("failed to cache user avatar; %w", err)
	}

	// Set admin cache (30 days)
	if err := rdb.Client.Set(ctx, adminKey, url, 30*24*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("failed to cache admin avatar; %w", err)
	}

	return url, nil
}

// Download remote image (user avatar)
func (u *User) refreshAvatar(
	ctx context.Context,
	config *config.Config,
	rdb *rdb.Service,
	r2s r2.Service) (string, error) {

	// Set the analytics ID in case it's missing
	if u.AnalyticsID == "" {
		u.SetAnalyticsID()
	}

	// Download the avatar from remote location
	data, err := u.downloadAvatar(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to download avatar; %w", err)
	}

	// Hash of the source image
	sourceHash := fmt.Sprintf("%x", md5.Sum(data)) // #nosec G401

	// Check for metadata - if we already have this version in R2
	head, err := r2s.HeadObject(
		ctx,
		config.R2CdnBucketName,
		fmt.Sprintf(avatarPath, u.AnalyticsID),
	)

	// Form the avatar URL
	avatarURL := &url.URL{
		Scheme:   "https",
		Host:     config.R2CdnDomain,
		Path:     fmt.Sprintf(avatarPath, u.AnalyticsID),
		RawQuery: "v=" + url.QueryEscape(sourceHash),
	}

	avatar := avatarURL.String()

	// If source unchanged just refresh both cache keys
	if err == nil && head.Metadata != nil {
		storedHash, exists := head.Metadata["source-hash"]
		if exists && storedHash == sourceHash {
			return u.updateAvatarCache(ctx, rdb, avatar)
		}
	}

	// Source changed or doesn't exist - decode and re-encode
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
		fmt.Sprintf(avatarPath, u.AnalyticsID),
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

	// Update both cache keys
	return u.updateAvatarCache(ctx, rdb, avatar)
}

// Delete local avatar if exists
func (u *User) DeleteAvatar(
	ctx context.Context,
	config *config.Config,
	rdb *rdb.Service,
	r2s r2.Service) {

	// Attemp to delete the avatar image from R2
	objectKey := fmt.Sprintf(avatarPath, u.AnalyticsID)
	if err := r2s.DeleteObject(ctx, config.R2CdnBucketName, objectKey); err != nil {
		log.Printf("Could not remove the avatar %s from R2: %v", objectKey, err)
	}

	// Delete all avatar Redis values
	for _, key := range []string{
		AvatarAdminPrefix + u.AnalyticsID,
		AvatarUserPrefix + u.AnalyticsID,
	} {
		if err := rdb.Client.Del(ctx, key).Err(); err != nil {
			log.Printf("Could not remove the avatar %s from Redis: %v", key, err)
		}
	}
}
