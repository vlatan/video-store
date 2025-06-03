package server

import (
	"factual-docs/internal/utils"
	"fmt"
	"io"
	"net/http"
	"os"
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
func (s *Server) downloadAvatar(avatarURL, analyticsID string) (filepath string, err error) {
	// Get remote file
	response, err := http.Get(avatarURL)
	if err != nil {
		return "", fmt.Errorf("can't read the remote file: %v", err)
	}
	defer response.Body.Close()

	// Create a file for writing
	path := fmt.Sprintf("%s/%s.jpg", s.config.RuntimeFileSystem, analyticsID)
	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("couldn't create file on disk: %v", err)
	}

	// Write to file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return "", fmt.Errorf("couldn't save the file on disk: %v", err)
	}

	return filepath, nil
}
