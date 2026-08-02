package posts

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"
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
func (s *Service) handleRate(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	var data struct {
		Rating uint8 `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to decode post rating",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	ratingData, err := s.postsRepo.Rate(r.Context(), data.Rating, userID, videoID)

	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}

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

	s.ui.WriteJSON(w, r, ratingData)
}

// Handle a post favorite from user
func (s *Service) handleReview(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	var data struct {
		Headline string `json:"headline"`
		Content  string `json:"content"`
		Rating   uint8  `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to decode post review",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if err := validateReview(data.Headline, data.Content); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to validate the review",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	reviewData, err := s.postsRepo.Review(
		r.Context(),
		userID,
		videoID,
		data.Rating,
		data.Headline,
		data.Content,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "user failed to review the video",
			"path", r.URL.Path,
			"userId", userID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	s.ui.WriteJSON(w, r, reviewData)
}
