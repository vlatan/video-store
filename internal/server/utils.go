package server

import (
	"factual-docs/internal/utils"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Extracts and sanitizes the value from the query param "redirect"
func getSafeRedirectPath(r *http.Request) string {
	redirectParam := r.URL.Query().Get("redirect")
	safePath, err := utils.SanitizeRelativePath(redirectParam)
	if err != nil {
		return "/"
	}
	return safePath
}

// Download remote image (user avatar)
func (s *Server) downloadAvatar(avatarURL, analyticsID string) (string, error) {
	// Get remote file
	response, err := http.Get(avatarURL)
	if err != nil {
		return "", fmt.Errorf("can't read the remote file: %v", err)
	}
	defer response.Body.Close()

	// Ensure the HTTP request was successful (status code 2xx)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf(
			"failed to download avatar from %s: received status code %d",
			avatarURL,
			response.StatusCode,
		)
	}

	// Create a file for writing
	destination := filepath.Join(s.config.DataVolume, analyticsID+".jpg")
	file, err := os.Create(destination)
	if err != nil {
		return "", fmt.Errorf("couldn't create file '%s': %v", destination, err)
	}

	// Flag to track if the download was successful
	valid := false

	// Run this clean up function on exit
	defer func() {
		if err := file.Close(); err != nil { // Close the file
			log.Printf("Warning: failed to close file '%s': %v\n", destination, err)
		}
		if !valid { // Remove the file if not successfuly created
			if err := os.Remove(destination); err != nil {
				log.Printf("Failed to remove partially created file '%s': %v\n", destination, err)
			}
		}
	}()

	// Write to file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't write to file '%s': %v", destination, err)
	}

	valid = true
	return destination, nil
}

func (s *Server) getLocalAvatar(r *http.Request, avatarURL, analyticsID string) string {
	// Get avatar URL from Redis
	redisKey := fmt.Sprintf("user:%s", analyticsID)
	avatar, err := s.rdb.Get(r.Context(), redisKey)
	if err == nil {
		return avatar
	}

	// Attempt to download the avatar, set default avatar on fail
	_, err = s.downloadAvatar(avatarURL, analyticsID)
	if err != nil {
		avatar = "/static/images/default-avatar.jpg"
	}

	// Save avatar URL to Redis and return
	avatar = "/static/images/avatars/" + analyticsID + ".jpg"
	s.rdb.Set(r.Context(), redisKey, avatar, 24*7*time.Hour)
	return avatar
}
