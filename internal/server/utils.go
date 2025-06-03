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

	// cachedData, err := s.rdb.Get(r.Context(), analyticsID)
	// if err != nil {

	// }

	// avatar, err := s.downloadAvatar(avatarURL, analyticsID)

	/*
		Check if redis key present
		If it is, do not attempt redownload:
			Check if there's avatar locally:
			If it's attach return the path
			If not return the default avatar path
		If it's not:
			Try to download the avatar
			Return downloaded avatar path
			If not return the default avatar path
		Set redis key with expiry
	*/

	return ""
}
