package store

import (
	"log/slog"
	"net/http"
)

// Flash message object to store to session for the next page
type FlashMessage struct {
	Message  string
	Category string
}

// AddFlash adds flash message in a session.
// No error returned if flashing fails.
func (s *Service) AddFlash(
	w http.ResponseWriter,
	r *http.Request,
	m *FlashMessage,
) {
	session, _ := s.Get(r, s.cfg.FlashSessionName)
	session.AddFlash(m)
	if err := session.Save(r, w); err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to save the flash session",
			"error", err,
		)
	}
}

// Flashes gets any flash messages from session
func (s *Service) Flashes(w http.ResponseWriter, r *http.Request) []*FlashMessage {

	// Get any flash messages from session
	session, _ := s.Get(r, s.cfg.FlashSessionName)
	flashes := session.Flashes()

	var flashMessages []*FlashMessage
	for _, v := range flashes {
		if flash, ok := v.(*FlashMessage); ok && flash != nil {
			flashMessages = append(flashMessages, flash)
		}
	}

	// Clear the flash session
	if err := s.Clear(r, w, session); err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to clear the flash session",
			"error", err,
		)
	}

	return flashMessages
}
