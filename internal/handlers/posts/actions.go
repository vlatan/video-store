package posts

import (
	"log"
	"net/http"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Handle a post like from user
func (s *Service) handleLike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Like(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not like the video %s: %v", userID, videoID, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
	}
}

// Handle a post unlike from user
func (s *Service) handleUnlike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Unlike(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not unlike the video %s: %v", userID, videoID, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
	}
}

// Handle a post favorite from user
func (s *Service) handleFave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Fave(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not fave the video %s: %v", userID, videoID, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
	}
}

// Handle a post unfavorite from user
func (s *Service) handleUnfave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Unfave(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not unfave the video %s: %v", userID, videoID, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
	}
}

// Handle a post ban
func (s *Service) handleBanPost(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.BanPost(r.Context(), videoID)
	if err != nil {
		log.Printf("User %d could not delete the video %s: %v", userID, videoID, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
		return
	}

	successDelete := models.FlashMessage{
		Message:  "The video has been deleted!",
		Category: "info",
	}

	s.ui.StoreFlashMessage(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
