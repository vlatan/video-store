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

// GetAvatar gets user avatar path, either from redis,
// or downloads and stores avatar path to redis.
// If the function returns an error the default avatar might be set.
func (u *User) GetAvatar(
	ctx context.Context,
	config *config.Config,
	rdb *rdb.Service,
	r2s r2.Service,
	keyPrefix string) error {

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

	// Refresh the avatar - reupload if changed
	avatar, err = u.refreshAvatar(ctx, config, r2s)

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	// Save non-nil error
	// Set to default avatar if no remote avatar
	if err != nil {
		avatar = defaultAvatar
		errs = append(errs, err)
	}

	// Set avatar
	u.LocalAvatarURL = avatar
	err = u.updateAvatarCache(ctx, rdb, avatar)

	// Return early if context error
	if utils.IsContextErr(err) {
		return err
	}

	errs = append(errs, err)
	return errors.Join(errs...)
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

// updateAvatarCache sets both user and admin avatar caches in Redis
func (u *User) updateAvatarCache(ctx context.Context, rdb *rdb.Service, url string) error {

	errs := make([]error, 0, 2)
	data := map[string]time.Duration{
		(AvatarUserPrefix + u.AnalyticsID):  24 * time.Hour,      // 1 day
		(AvatarAdminPrefix + u.AnalyticsID): 30 * 24 * time.Hour, // 30 days
	}

	for key, ttl := range data {
		err := rdb.Client.Set(ctx, key, url, ttl).Err()
		errs = append(errs, err)
	}

	return errors.Join(errs...)
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

	// Hash the source image
	sourceHash := fmt.Sprintf("%x", md5.Sum(data)) // #nosec G401

	// Get R2 object head
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
	objectKey := fmt.Sprintf(avatarPath, u.AnalyticsID)
	err := r2s.DeleteObject(ctx, config.R2CdnBucketName, objectKey)
	err = fmt.Errorf("failed to remove avatar %q from R2: %w", objectKey, err)
	errs = append(errs, err)

	// Delete user and admin avatar Redis cache values
	for _, key := range []string{
		AvatarAdminPrefix + u.AnalyticsID,
		AvatarUserPrefix + u.AnalyticsID,
	} {
		err := rdb.Client.Del(ctx, key).Err()
		err = fmt.Errorf("failed to remove avatar %q from Redis: %w", key, err)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
