package handler

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
func (s *Service) ActionAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"like", "fave", "rate", "review"}
	if !slices.Contains(allowedActions, action) {
		slog.WarnContext(r.Context(), "not a valid action on post")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the current user
	user := ctxd.GetUser(r.Context())

	switch action {
	case "like":
		s.like(w, r, user.ID, videoID)
	case "fave":
		s.fave(w, r, user.ID, videoID)
	case "rate":
		s.rate(w, r, user.ID, videoID)
	case "review":
		s.review(w, r, user.ID, videoID)
	default:
		s.tmpl.JSONError(w, r, http.StatusBadRequest)
	}
}

// UndoPostActionAPI undoes action on a video
func (s *Service) UndoActionAPI(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"unlike", "unfave", "unrate"}
	if !slices.Contains(allowedActions, action) {
		slog.WarnContext(r.Context(), "not a valid action on post")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the current user
	user := ctxd.GetUser(r.Context())

	switch action {
	case "unlike":
		s.unlike(w, r, user.ID, videoID)
	case "unfave":
		s.unfave(w, r, user.ID, videoID)
	case "unrate":
		s.unrate(w, r, user.ID, videoID)
	default:
		s.tmpl.JSONError(w, r, http.StatusBadRequest)
	}
}

// Handle a post like from user
func (s *Service) like(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	rowsAffected, err := s.post.Like(r.Context(), userID, videoID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to like the video",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such video to like")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
	}
}

// Handle a post unlike from user
func (s *Service) unlike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	rowsAffected, err := s.post.Unlike(r.Context(), userID, videoID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to unlike the video",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such video to unlike")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
	}
}

// Handle a post favorite from user
func (s *Service) fave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	rowsAffected, err := s.post.Fave(r.Context(), userID, videoID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to fave the video",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such video to fave")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
	}
}

// Handle a post unfavorite from user
func (s *Service) unfave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	rowsAffected, err := s.post.Unfave(r.Context(), userID, videoID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to unfave the video",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such video to unfave")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
	}
}

// Handle a post favorite from user
func (s *Service) rate(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	var data struct {
		Rating uint8 `json:"rating"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		slog.ErrorContext(
			r.Context(), "failed to decode post rating",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if data.Rating < 1 || data.Rating > 10 {
		slog.WarnContext(
			r.Context(), "rating out of bounds",
		)
		s.tmpl.JSONError(w, r, http.StatusBadRequest)
		return
	}

	ratingStats, err := s.post.Rate(r.Context(), data.Rating, userID, videoID)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(r.Context(), "no such video to rate")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to rate the video",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	s.tmpl.WriteJSON(w, r, ratingStats)
}

// handleUnrate handles a deletion of a user rating/review
func (s *Service) unrate(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

	ratingStats, err := s.post.Unrate(r.Context(), userID, videoID)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(r.Context(), "no such video to unrate")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to unrate the video",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	s.tmpl.WriteJSON(w, r, ratingStats)
}

// Handle a post favorite from user
func (s *Service) review(w http.ResponseWriter, r *http.Request, userID int, videoID string) {

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
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if data.Rating < 1 || data.Rating > 10 {
		slog.WarnContext(r.Context(), "rating out of bounds")
		s.tmpl.JSONError(w, r, http.StatusBadRequest)
		return
	}

	if err := validateReview(data.Headline, data.Content); err != nil {
		slog.WarnContext(
			r.Context(), "failed to validate the review",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusBadRequest)
		return
	}

	reviewData, err := s.post.Review(
		r.Context(),
		userID,
		videoID,
		data.Rating,
		data.Headline,
		data.Content,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(r.Context(), "no such video to review")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to review the video",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	s.tmpl.WriteJSON(w, r, reviewData)
}
