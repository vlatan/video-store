package posts

import (
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/utils"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Handle the Home page
func (s *Service) HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Generate template data
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetCurrentUser(w, r)

	// Get page number from a query param
	page := utils.GetPageNum(r)
	redisKey := fmt.Sprintf("posts:page:%d", page)

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	var posts []database.Post
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		&posts,
		func() ([]database.Post, error) {
			return s.db.GetPosts(r.Context(), page, orderBy)
		},
	)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		if page > 1 {
			s.tm.JSONError(w, r, http.StatusInternalServerError)
			return
		}
		s.tm.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if len(posts) == 0 {
		log.Printf("Fetched zero posts on URI '%s'", r.RequestURI)
		if page > 1 {
			s.tm.JSONError(w, r, http.StatusNotFound)
			return
		}
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	// If not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		s.tm.WriteJSON(w, r, posts)
		return
	}

	data.Posts.Items = posts
	s.tm.RenderHTML(w, r, "home", data)
}
