package users

import (
	"factual-docs/internal/shared/utils"
	"log"
	"net/http"
	"time"
)

// Handle the user favorites page
func (s *Service) UserFavoritesHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page number from the request query param
	page := utils.GetPageNum(r)

	// Generate template data
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

	posts, err := s.postsRepo.GetUserGavedPosts(r.Context(), data.CurrentUser.ID, page)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		if page > 1 {
			s.tm.JSONError(w, r, http.StatusInternalServerError)
			return
		}
		s.tm.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if page > 1 && len(posts.Items) == 0 {
		log.Printf("Fetched zero posts on URI '%s'", r.RequestURI)
		s.tm.JSONError(w, r, http.StatusNotFound)
		return
	}

	// If not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		s.tm.WriteJSON(w, r, posts.Items)
		return
	}

	data.Posts = posts
	data.Title = "Your Favorite Documentaries:"
	s.tm.RenderHTML(w, r, "user_library.html", data)
}
