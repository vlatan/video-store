package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"factual-docs/internal/database"
	"factual-docs/internal/redis"
	"factual-docs/internal/templates"
	"factual-docs/internal/utils"
	"factual-docs/web"
	"fmt"
	"log"
	"maps"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/markbates/goth/gothic"
)

var validVideoID = regexp.MustCompile("^([-a-zA-Z0-9_]{11})$")

// Handle the Home page
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {

	// Generate template data
	data := s.NewData(w, r)

	// Get page number from a query param
	page := getPageNum(r)
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
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if len(posts) == 0 {
		http.NotFound(w, r)
		return
	}

	// if not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		if err := s.tm.WriteJSON(w, posts); err != nil {
			log.Println(err)
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		}
		return
	}

	data.Posts.Items = posts
	if err := s.tm.Render(w, "home", data); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}
}

func (s *Server) categoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	// Get category slug from URL
	slug := r.PathValue("category")

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := s.NewData(w, r)

	// Check if the category is valid
	category, valid := isValidCategory(data.Categories, slug)
	if !valid {
		http.NotFound(w, r)
		return
	}

	// Get page number from a query param
	page := getPageNum(r)
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
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if len(posts) == 0 {
		http.NotFound(w, r)
		return
	}

	// if not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		if err := s.tm.WriteJSON(w, posts); err != nil {
			log.Println(err)
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		}
		return
	}

	data.Posts.Items = posts
	data.Title = category.Name
	if err := s.tm.Render(w, "category", data); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}
}

// Handle the requests from the searchform
func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	// Get the search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get the page number from the request query param
	page := getPageNum(r)

	// Generate the default data
	data := s.NewData(w, r)
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
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if page > 1 && len(posts.Items) == 0 {
		http.NotFound(w, r)
		return
	}

	// if not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		if err := s.tm.WriteJSON(w, posts.Items); err != nil {
			log.Println(err)
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		}
		return
	}
	data.Posts = posts
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())
	if err := s.tm.Render(w, "search", data); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}
}

// Handle a single post
func (s *Server) singlePostHandler(w http.ResponseWriter, r *http.Request) {
	// Get category slug from URL
	videoID := r.PathValue("video")

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		log.Println("Not a valid video ID:", videoID)
		http.NotFound(w, r)
		return
	}

	// Generate the default data
	data := s.NewData(w, r)

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
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf("Error while getting the video '%s' from DB: %v\n", videoID, err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if post.ID == 0 {
		log.Println("Can't find the video in DB:", videoID)
		http.NotFound(w, r)
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
	if err := s.tm.Render(w, "post", data); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}
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
		http.NotFound(w, r)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"like", "unlike", "fave", "unfave", "edit", "delete"}
	if !slices.Contains(allowedActions, action) {
		log.Printf("Not a valid action '%s' on video: %s\n", action, videoID)
		http.NotFound(w, r)
		return
	}

	// Check if the user is authenticated
	currentUser := s.getCurrentUser(w, r)
	if !currentUser.IsAuthenticated() {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	// Check if user is authorized to edit or delete
	if (action == "edit" || action == "delete") && currentUser.UserID != s.config.AdminOpenID {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
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
		s.handleDelete(w, r, currentUser.ID, videoID)
	default:
		http.Error(w, "Invalid action", http.StatusBadRequest)
	}
}

// Handle the title or description update
func (s *Server) handleEdit(w http.ResponseWriter, r *http.Request, videoID string, currentUser *templates.User) {
	var data bodyData

	// Deocode JSON
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Could not decode the JSON body on path: %s", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Check for title or description
	if data.Title == "" && data.Description == "" {
		log.Printf("No title and description in body on path: %s", r.URL.Path)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
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

// Handle minified static file from cache
func (s *Server) staticHandler(w http.ResponseWriter, r *http.Request) {

	// VERY IMPORTANT: Do not allow directory browsing
	if strings.HasSuffix(r.URL.Path, "/") {
		http.NotFound(w, r)
		return
	}

	// Set long max age cache conttrol
	w.Header().Set("Cache-Control", "max-age=31536000")

	// Get the file information
	fileInfo, ok := s.sf[r.URL.Path]

	// Set content type header if media type available
	if ok && fileInfo.MediaType != "" {
		w.Header().Set("Content-Type", fileInfo.MediaType)
	}

	// Set Etag if etag available
	if ok && fileInfo.Etag != "" {
		w.Header().Set("Etag", fileInfo.Etag)
	}

	// Serve the file content if we have bytes stored
	if ok && fileInfo.Bytes != nil && len(fileInfo.Bytes) > 0 {
		http.ServeContent(w, r, r.URL.Path, time.Time{}, bytes.NewReader(fileInfo.Bytes))
		return
	}

	// Sanitize the path
	name, err := utils.SanitizeRelativePath(r.URL.Path)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Serve user avatars from the data volume
	if strings.HasPrefix(name, "/static/images/avatars/") {
		parsed, err := url.Parse(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		name = s.config.DataVolume + "/" + filepath.Base(parsed.Path)
		http.ServeFile(w, r, name)
		return
	}

	// Try to serve from the embedded FS
	http.ServeFileFS(w, r, web.Files, name)
}

// DB and Redis health status
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {

	dbStatus := s.db.Health(r.Context())
	rdbStatus := s.rdb.Health(r.Context())

	maps.Copy(dbStatus, rdbStatus)

	status, err := json.Marshal(dbStatus)
	if err != nil {
		http.Error(w,
			"Failed to marshal health check response",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(status); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

// Provider Auth
func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getSafeRedirectPath(r)

	// Auth with gothic, try to get the user without re-authenticating
	gothUser, err := gothic.CompleteUserAuth(w, r)

	// If unable to re-auth start the auth from the beginning
	if err != nil {
		// Store this redirect URL in another session as flash message
		session, _ := s.store.Get(r, s.config.FlashSessionName)
		session.AddFlash(redirectTo, "redirect")
		session.Save(r, w)

		// Begin Provider auth
		// This will redirect the client to the provider's authentication end-point
		gothic.BeginAuthHandler(w, r)
		return
	}

	// Login user, save into our session
	if err = s.loginUser(w, r, &gothUser); err != nil {
		log.Printf("Error logging in the user: %v", err)
		s.storeFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.storeFlashMessage(w, r, &successLogin)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Provider Auth callback
func (s *Server) authCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := s.getUserFinalRedirect(w, r)

	// Authenticate the user using gothic
	gothUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		log.Printf("Error with gothic user auth: %v", err)
		s.storeFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Save user into our session
	if err = s.loginUser(w, r, &gothUser); err != nil {
		log.Printf("Error logging in the user: %v", err)
		s.storeFlashMessage(w, r, &failedLogin)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.storeFlashMessage(w, r, &successLogin)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

// Logout user, delete sessions
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getSafeRedirectPath(r)

	// Redirect to home if user is not logged in
	if user := s.getCurrentUser(w, r); user == nil || user.UserID == "" {
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Remove gothic session if any
	if err := gothic.Logout(w, r); err != nil {
		log.Printf("Error loging out the user with gothic: %v", err)
		s.storeFlashMessage(w, r, &failedLogout)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Remove user's session
	if err := s.logoutUser(w, r); err != nil {
		log.Printf("Error loging out the user: %v", err)
		s.storeFlashMessage(w, r, &failedLogout)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	s.storeFlashMessage(w, r, &successLogout)
	http.Redirect(w, r, redirectTo, http.StatusSeeOther)
}

func isValidCategory(categories []database.Category, slug string) (database.Category, bool) {
	for _, category := range categories {
		if category.Slug == slug {
			return category, true
		}
	}
	return database.Category{}, false
}

// Get page number from the request query param
// Defaults to 1 if invalid param
func getPageNum(r *http.Request) int {
	page := 1
	pageStr := r.URL.Query().Get("page")
	pageInt, err := strconv.Atoi(pageStr)
	if err == nil || pageInt != 0 {
		page = pageInt
	}
	return page
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
