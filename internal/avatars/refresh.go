package avatars

import (
	"bytes"
	"context"
	"crypto/md5" // #nosec G501
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/url"

	"github.com/vlatan/video-store/internal/models"
)

// refreshAvatar reuploads the user avatar at R2 if changed
func (s *Service) refreshAvatar(ctx context.Context, user *models.User) (string, error) {

	// Download the avatar from remote location
	data, err := s.downloadAvatar(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to download avatar; %w", err)
	}

	// Hash the source image.
	// We are not doing anything important here, md5  is fine.
	sourceHash := fmt.Sprintf("%x", md5.Sum(data)) // #nosec G401

	// Get R2 object head
	head, err := s.r2s.HeadObject(
		ctx,
		s.config.R2CdnBucketName,
		fmt.Sprintf(avatarR2Path, user.PublicID),
	)

	// Form the avatar URL
	avatarURL := &url.URL{
		Scheme:   "https",
		Host:     s.config.R2CdnDomain,
		Path:     fmt.Sprintf(avatarR2Path, user.PublicID),
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
			user.AvatarURL, err,
		)
	}

	// Convert to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf(
			"failed to convert the avatar %s to JPEG; %w",
			user.AvatarURL, err,
		)
	}

	// Upload object to bucket
	err = s.r2s.PutObject(
		ctx,
		s.config.R2CdnBucketName,
		fmt.Sprintf(avatarR2Path, user.PublicID),
		bytes.NewReader(buf.Bytes()),
		"image/jpeg",
		map[string]string{"source-hash": sourceHash},
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to upload the avatar %s to bucket: %w",
			user.PublicID, err,
		)
	}

	// Return the avatar
	return avatar, nil
}

// downloadAvatar downloads avatar from a remote source
func (s *Service) downloadAvatar(ctx context.Context, user *models.User) ([]byte, error) {

	// Create a request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, user.AvatarURL, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create GET request for avatar %q download; %w",
			user.AvatarURL, err,
		)
	}

	// Execute the request
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to execute GET request for avatar %q; %w",
			user.AvatarURL, err,
		)
	}
	defer resp.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"received status code %d for avatar %q",
			resp.StatusCode, user.AvatarURL,
		)
	}

	// Limit size to prevent abuse (5MB max)
	limitedReader := io.LimitReader(resp.Body, 5*1024*1024)
	return io.ReadAll(limitedReader)
}
