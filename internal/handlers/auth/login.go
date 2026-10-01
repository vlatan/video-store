package auth

import (
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/types"
)

// Store user info in our own session
func (s *Service) loginUser(w http.ResponseWriter, r *http.Request, user *types.User) error {

	// Make user public ID
	var err error
	user.PublicID, err = user.MakePublicID()
	if err != nil {
		return err
	}

	// Update or insert user
	user.ID, err = s.usersRepo.UpsertUser(r.Context(), user)
	if err != nil {
		return err
	}
	user.LastSeen = new(time.Now())

	// Add user to session
	if err := s.session.AddUser(w, r, user); err != nil {
		return err
	}

	// Download and save the avatar if not in Redis cache
	if err := s.avatar.Save(r.Context(), user); err != nil {
		return err
	}

	return nil
}
