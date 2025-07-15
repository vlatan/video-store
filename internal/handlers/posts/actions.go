package posts

import (
	"encoding/json"
	"factual-docs/internal/models"
	"log"
	"net/http"
)

type bodyData struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Handle a post like from user
func (s *Service) handleLike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Like(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not like the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to like.\n", videoID)
		http.NotFound(w, r)
	}
}

// Handle a post unlike from user
func (s *Service) handleUnlike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Unlike(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not unlike the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to unlike.\n", videoID)
		http.NotFound(w, r)
	}
}

// Handle a post favorite from user
func (s *Service) handleFave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Fave(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not fave the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to fave.\n", videoID)
		http.NotFound(w, r)
	}
}

// Handle a post unfavorite from user
func (s *Service) handleUnfave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.Unfave(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not unfave the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to unfave.\n", videoID)
		http.NotFound(w, r)
	}
}

// Handle a post title update
func (s *Service) handleUpdateTitle(w http.ResponseWriter, r *http.Request, userID int, videoID, title string) {
	rowsAffected, err := s.postsRepo.UpdateTitle(r.Context(), videoID, title)
	if err != nil {
		log.Printf("User %d could not update the title of the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to update the title of.\n", videoID)
		http.NotFound(w, r)
	}
}

// Handle a post description update
func (s *Service) handleUpdateDesc(w http.ResponseWriter, r *http.Request, userID int, videoID, description string) {
	rowsAffected, err := s.postsRepo.UpdateDesc(r.Context(), videoID, description)
	if err != nil {
		log.Printf("User %d could not update the description of the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to update the description of.\n", videoID)
		http.NotFound(w, r)
	}
}

// Handle a post description update
func (s *Service) handleDeletePost(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.postsRepo.DeletePost(r.Context(), videoID)
	if err != nil {
		log.Printf("User %d could not delete the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to delete.\n", videoID)
		http.NotFound(w, r)
		return
	}

	successDelete := models.FlashMessage{
		Message:  "The video has been deleted!",
		Category: "info",
	}

	s.tm.StoreFlashMessage(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handle the title or description update
func (s *Service) handleEdit(w http.ResponseWriter, r *http.Request, videoID string, currentUser *models.User) {
	var data bodyData

	// Deocode JSON
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Could not decode the JSON body on path: %s", r.URL.Path)
		s.tm.JSONError(w, r, http.StatusBadRequest)
		return
	}

	// Check for title or description
	if data.Title == "" && data.Description == "" {
		log.Printf("No title and description in body on path: %s", r.URL.Path)
		s.tm.JSONError(w, r, http.StatusBadRequest)
		return
	}

	if data.Title != "" {
		s.handleUpdateTitle(w, r, currentUser.ID, videoID, data.Title)
		return
	}

	if data.Description != "" {
		s.handleUpdateDesc(w, r, currentUser.ID, videoID, data.Description)
	}
}
