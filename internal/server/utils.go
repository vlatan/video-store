package server

import (
	"factual-docs/internal/utils"
	"fmt"
	"io"
	"net/http"
	"os"
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

	// Create a file for writing
	path := fmt.Sprintf("%s/%s.jpg", s.config.DataVolume, analyticsID)
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("couldn't create file on disk: %v", err)
	}

	// Write to file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't save the file on disk: %v", err)
	}

	return path, nil
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
