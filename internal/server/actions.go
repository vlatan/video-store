package server

import (
	"factual-docs/internal/shared/database"
	tmpls "factual-docs/internal/shared/templates"
	"log"
	"net/http"
)

// Handle a post like from user
func (s *Server) handleLike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.db.Exec(r.Context(), database.LikeQuery, userID, videoID)
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
func (s *Server) handleUnlike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.db.Exec(r.Context(), database.UnlikeQuery, userID, videoID)
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
func (s *Server) handleFave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.db.Exec(r.Context(), database.FaveQuery, userID, videoID)
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
func (s *Server) handleUnfave(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.db.Exec(r.Context(), database.UnfaveQuery, userID, videoID)
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
func (s *Server) handleUpdateTitle(w http.ResponseWriter, r *http.Request, userID int, videoID, title string) {
	rowsAffected, err := s.db.Exec(r.Context(), database.UpdateTitleQuery, videoID, title)
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
func (s *Server) handleUpdateDesc(w http.ResponseWriter, r *http.Request, userID int, videoID, description string) {
	rowsAffected, err := s.db.Exec(r.Context(), database.UpdateDescQuery, videoID, description)
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
func (s *Server) handleDeletePost(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.db.Exec(r.Context(), database.DeletePostQuery, videoID)
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

	successDelete := tmpls.FlashMessage{
		Message:  "The video has been deleted!",
		Category: "info",
	}

	s.auth.StoreFlashMessage(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
