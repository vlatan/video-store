package avatar

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

	_ "image/gif" // Register GIF decoder
	_ "image/png" // Register PNG decoder

	_ "golang.org/x/image/webp" // Register WebP decoder
)

const sourceHashKey = "source-hash"

// refreshAvatar reuploads the user avatar at R2 if changed
func (s *Service) refresh(ctx context.Context, publicId, uri string) (string, error) {

	// Download the avatar from remote location
	data, err := s.download(ctx, uri)
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
		fmt.Sprintf(avatarR2Path, publicId),
	)

	// Form the avatar URL
	avatarURL := &url.URL{
		Scheme:   "https",
		Host:     s.config.R2CdnDomain,
		Path:     fmt.Sprintf(avatarR2Path, publicId),
		RawQuery: "v=" + url.QueryEscape(sourceHash),
	}

	avatar := avatarURL.String()

	// If object source unchanged avatar not changed, return it
	if err == nil && head.Metadata != nil {
		storedHash, exists := head.Metadata[sourceHashKey]
		if exists && storedHash == sourceHash {
			return avatar, nil
		}
	}

	// Decode the avatar
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf(
			"failed to decode the file for avatar %s; %w",
			uri, err,
		)
	}

	// Convert to JPEG
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	if err != nil {
		return "", fmt.Errorf(
			"failed to convert the avatar %s to JPEG; %w",
			uri, err,
		)
	}

	// Upload object to bucket
	err = s.r2s.PutObject(
		ctx,
		s.config.R2CdnBucketName,
		fmt.Sprintf(avatarR2Path, publicId),
		bytes.NewReader(buf.Bytes()),
		"image/jpeg",
		map[string]string{sourceHashKey: sourceHash},
	)

	if err != nil {
		return "", fmt.Errorf(
			"failed to upload the avatar %s to bucket: %w",
			publicId, err,
		)
	}

	// Return the avatar
	return avatar, nil
}

// downloadAvatar downloads avatar from a remote source
func (s *Service) download(ctx context.Context, uri string) ([]byte, error) {

	// Create a request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create GET request for avatar %q download; %w",
			uri, err,
		)
	}

	// Execute the request
	var client = &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to execute GET request for avatar %q; %w",
			uri, err,
		)
	}
	defer resp.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"received status code %d for avatar %q",
			resp.StatusCode, uri,
		)
	}

	// Limit size to prevent abuse (5MB max)
	limitedReader := io.LimitReader(resp.Body, 5*1024*1024)
	return io.ReadAll(limitedReader)
}
