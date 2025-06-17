package server

import (
	"log"
	"net/http"
)

func (s *Server) handleLike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.db.Like(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not like the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to like.\n", videoID)
		http.NotFound(w, r)
		return
	}
}

func (s *Server) handleUnlike(w http.ResponseWriter, r *http.Request, userID int, videoID string) {
	rowsAffected, err := s.db.Unlike(r.Context(), userID, videoID)
	if err != nil {
		log.Printf("User %d could not unlike the video %s: %v\n", userID, videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such video %s to unlike.\n", videoID)
		http.NotFound(w, r)
		return
	}
}
