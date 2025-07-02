package server

import (
	"context"
	"encoding/json"
	"errors"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/redis"
	tmpls "factual-docs/internal/shared/templates"
	"factual-docs/internal/utils"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

var validVideoID = regexp.MustCompile("^([-a-zA-Z0-9_]{11})$")

// Handle the requests from the searchform
func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
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
		s.tm.WriteJSON(w, r, posts)
		return
	}

	data.Posts = posts
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())
	s.tm.RenderHTML(w, r, "search", data)
}

// Handle a single post
func (s *Server) singlePostHandler(w http.ResponseWriter, r *http.Request) {
	// Get category slug from URL
	videoID := r.PathValue("video")

	// Generate the default data
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetCurrentUser(w, r)

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		log.Println("Not a valid video ID:", videoID)
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	var post database.Post
	err := redis.GetItems(
		!data.CurrentUser.IsAuthenticated(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("post:%s", videoID),
		s.config.CacheTimeout,
		&post,
		func() (database.Post, error) {
			return s.db.GetSinglePost(r.Context(), videoID)
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Println("Can't find the video in DB:", videoID)
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	if err != nil {
		log.Printf("Error while getting the video '%s' from DB: %v", videoID, err)
		s.tm.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if post.ID == 0 {
		log.Println("Can't find the video in DB:", videoID)
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	// Assign the post to data
	data.CurrentPost = &post
	data.Title = post.Title

	// Check whether the current user liked and/or faved the post
	if data.CurrentUser.IsAuthenticated() {
		userActions, _ := s.db.GetUserActions(
			r.Context(),
			data.CurrentUser.ID,
			data.CurrentPost.ID,
		)
		data.CurrentPost.CurrentUserLiked = userActions.Liked
		data.CurrentPost.CurrentUserFaved = userActions.Faved
	}

	// Ignore the error on related posts, no posts will be shown
	var relatedPosts []database.Post
	redis.GetItems(
		!data.CurrentUser.IsAuthenticated(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("post:%s:related_posts", videoID),
		s.config.CacheTimeout,
		&relatedPosts,
		func() ([]database.Post, error) {
			return s.getRelatedPosts(r.Context(), post.Title)
		},
	)

	data.CurrentPost.RelatedPosts = relatedPosts
	s.tm.RenderHTML(w, r, "post", data)
}

type bodyData struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Perform an action on a video
func (s *Server) postActionHandler(w http.ResponseWriter, r *http.Request) {

	// This is a post request, close the body on exit
	defer r.Body.Close()

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		log.Println("Not a valid video ID:", videoID)
		s.tm.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"like", "unlike", "fave", "unfave", "edit", "delete"}
	if !slices.Contains(allowedActions, action) {
		log.Printf("Not a valid action '%s' on video: %s\n", action, videoID)
		s.tm.JSONError(w, r, http.StatusNotFound)
		return
	}

	// Get the current user
	currentUser := s.auth.GetCurrentUser(w, r)

	// Check if user is authorized to edit or delete (admin)
	if (action == "edit" || action == "delete") &&
		currentUser.UserID != s.config.AdminOpenID {
		s.tm.JSONError(w, r, http.StatusForbidden)
		return
	}

	switch action {
	case "like":
		s.handleLike(w, r, currentUser.ID, videoID)
	case "unlike":
		s.handleUnlike(w, r, currentUser.ID, videoID)
	case "fave":
		s.handleFave(w, r, currentUser.ID, videoID)
	case "unfave":
		s.handleUnfave(w, r, currentUser.ID, videoID)
	case "edit":
		s.handleEdit(w, r, videoID, currentUser)
	case "delete":
		s.handleDeletePost(w, r, currentUser.ID, videoID)
	default:
		s.tm.JSONError(w, r, http.StatusBadRequest)
	}
}

// Handle the title or description update
func (s *Server) handleEdit(w http.ResponseWriter, r *http.Request, videoID string, currentUser *tmpls.User) {
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

// Handle ads.txt
func (s *Server) adsTextHandler(w http.ResponseWriter, r *http.Request) {
	if s.config.AdSenseAccount == "" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	content := fmt.Sprintf("google.com, pub-%s, DIRECT, f08c47fec0942fa0", s.config.AdSenseAccount)
	if _, err := w.Write([]byte(content)); err != nil {
		log.Printf("Failed to write response to '/ads.txt': %v", err)
	}
}

// DB and Redis health status
// Wrap this with middlware that allows only admins
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {

	// Construct joined map
	data := map[string]any{
		"redis_status":    s.rdb.Health(r.Context()),
		"database_status": s.db.Health(r.Context()),
		"server_status":   getServerStats(),
	}

	s.tm.WriteJSON(w, r, data)
}

// Get post's related posts based on provided title as search query
func (s *Server) getRelatedPosts(ctx context.Context, title string) (posts []database.Post, err error) {
	// Search the DB for posts
	searchedPosts, err := s.db.SearchPosts(ctx, title, s.config.NumRelatedPosts+1, 0)

	if err != nil {
		return posts, err
	}

	for _, sp := range searchedPosts.Items {
		if sp.Title != title {
			posts = append(posts, sp)
		}
	}

	return posts, err
}
