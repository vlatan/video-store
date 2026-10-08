package posts

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/vlatan/video-store/internal/ctxd"
)

// PostActionAPI performs POST action on a video
func (s *Service) PostActionAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"like", "fave", "rate", "review"}
	if !slices.Contains(allowedActions, action) {
		slog.WarnContext(r.Context(), "not a valid action on post")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the current user
	user := ctxd.GetUser(r.Context())

	switch action {
	case "like":
		s.handleLike(w, r, user.ID, videoID)
	case "fave":
		s.handleFave(w, r, user.ID, videoID)
	case "rate":
		s.handleRate(w, r, user.ID, videoID)
	case "review":
		s.handleReview(w, r, user.ID, videoID)
	default:
		s.ui.JSONError(w, r, http.StatusBadRequest)
	}
}

// UndoPostActionAPI undoes action on a video
func (s *Service) UndoPostActionAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"unlike", "unfave", "unrate"}
	if !slices.Contains(allowedActions, action) {
		slog.WarnContext(r.Context(), "not a valid action on post")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the current user
	user := ctxd.GetUser(r.Context())

	switch action {
	case "unlike":
		s.handleUnlike(w, r, user.ID, videoID)
	case "unfave":
		s.handleUnfave(w, r, user.ID, videoID)
	case "unrate":
		s.handleUnrate(w, r, user.ID, videoID)
	default:
		s.ui.JSONError(w, r, http.StatusBadRequest)
	}
}

// Handle a post like from user
func (s *Service) handleLike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	rowsAffected, err := s.postsRepo.Like(r.Context(), userID, videoID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to like the video",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such video to like")
		s.ui.JSONError(w, r, http.StatusNotFound)
	}
}

// Handle a post unlike from user
func (s *Service) handleUnlike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	rowsAffected, err := s.postsRepo.Unlike(r.Context(), userID, videoID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to unlike the video",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such video to unlike")
		s.ui.JSONError(w, r, http.StatusNotFound)
	}
}

// Handle a post favorite from user
func (s *Service) handleFave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	rowsAffected, err := s.postsRepo.Fave(r.Context(), userID, videoID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to fave the video",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such video to fave")
		s.ui.JSONError(w, r, http.StatusNotFound)
	}
}

// Handle a post unfavorite from user
func (s *Service) handleUnfave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	rowsAffected, err := s.postsRepo.Unfave(r.Context(), userID, videoID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to unfave the video",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such video to unfave")
		s.ui.JSONError(w, r, http.StatusNotFound)
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
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if data.Rating < 1 || data.Rating > 10 {
		slog.WarnContext(
			r.Context(), "rating out of bounds",
		)
		s.ui.JSONError(w, r, http.StatusBadRequest)
		return
	}

	ratingStats, err := s.postsRepo.Rate(r.Context(), data.Rating, userID, videoID)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(r.Context(), "no such video to rate")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to rate the video",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	s.ui.WriteJSON(w, r, ratingStats)
}

// handleUnrate handles a deletion of a user rating/review
func (s *Service) handleUnrate(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	ratingStats, err := s.postsRepo.Unrate(r.Context(), userID, videoID)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(r.Context(), "no such video to unrate")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to unrate the video",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	s.ui.WriteJSON(w, r, ratingStats)
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
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if data.Rating < 1 || data.Rating > 10 {
		slog.WarnContext(r.Context(), "rating out of bounds")
		s.ui.JSONError(w, r, http.StatusBadRequest)
		return
	}

	if err := validateReview(data.Headline, data.Content); err != nil {
		slog.WarnContext(
			r.Context(), "failed to validate the review",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusBadRequest)
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
		slog.WarnContext(r.Context(), "no such video to review")
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to review the video",
			"error", err,
		)
		s.ui.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	s.ui.WriteJSON(w, r, reviewData)
}
