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

// Handle posts in a certain category
func (s *Service) CategoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	// Get category slug from URL
	slug := r.PathValue("category")

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetCurrentUser(w, r)

	// Check if the category is valid
	category, valid := isValidCategory(data.Categories, slug)
	if !valid {
		log.Printf("Asked for invalid category '%s' on URI '%s'", slug, r.RequestURI)
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	// Get page number from a query param
	page := utils.GetPageNum(r)
	redisKey := fmt.Sprintf("%s:posts:page:%d", slug, page)

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
			return s.db.GetCategoryPosts(r.Context(), slug, orderBy, page)
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

	// if not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		s.tm.WriteJSON(w, r, posts)
		return
	}

	data.Posts.Items = posts
	data.Title = category.Name
	s.tm.RenderHTML(w, r, "category", data)
}

// Handle the requests from the searchform
func (s *Service) SearchPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get the page number from the request query param
	page := utils.GetPageNum(r)

	// Generate the default data
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetCurrentUser(w, r)
	data.SearchQuery = searchQuery

	limit := s.config.PostsPerPage
	offset := (page - 1) * limit

	start := time.Now()
	encodedSearchQuery := utils.EscapeTrancateString(searchQuery, 100)

	// For search posts we are using the database.Posts struct,
	// so we can add total results and time took
	var posts database.Posts
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("posts:search:page:%d:%s", page, encodedSearchQuery),
		s.config.CacheTimeout,
		&posts,
		func() (database.Posts, error) {
			return s.db.SearchPosts(r.Context(), searchQuery, limit, offset)
		},
	)

	end := time.Since(start)

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
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())
	s.tm.RenderHTML(w, r, "search", data)
}
