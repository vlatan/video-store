package posts

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Handle a post like from user
func (s *Service) handleLike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Like(r.Context(), userID, videoID)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "user failed to like the video",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
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
		slog.ErrorContext(
			r.Context(), "user failed to unlike the video",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
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
		slog.ErrorContext(
			r.Context(), "user failed to favorite the video",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
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
		slog.ErrorContext(
			r.Context(), "user failed to unfavorite the video",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
	}
}

// Handle a post favorite from user
func (s *Service) handleRate(w http.ResponseWriter, r *http.Request, rating, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Rate(r.Context(), rating, userID, videoID)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "user failed to rate the video",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
	}
}

// Handle a post ban
func (s *Service) handleBan(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.BanPost(r.Context(), videoID)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "user failed to ban/delete the video",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
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
