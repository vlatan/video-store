package server

import (
	"bytes"
	"context"
	"encoding/json"
	"factual-docs/internal/database"
	"factual-docs/internal/redis"
	"factual-docs/internal/utils"
	"factual-docs/web"
	"fmt"
	"log"
	"maps"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/markbates/goth/gothic"
)

// Handle the Home page
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {

	page := 1
	pageStr := r.URL.Query().Get("page")
	pageInt, err := strconv.Atoi(pageStr)
	if err == nil || pageInt != 0 {
		page = pageInt
	}

	var posts []database.Post

	// Use the generic cache wrapper
	err = redis.Cached(
		r.Context(),
		s.rdb,
		fmt.Sprintf("posts_page_%d", page),
		24*time.Hour,
		&posts,
		func() ([]database.Post, error) {
			return s.db.GetPosts(page) // Call the actual underlying database method
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

	data := s.NewData(w, r)
	data.Posts = posts

	if err := s.tm.Render(w, "home", data); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
	}
}

func (s *Server) categoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	page := 1
	pageStr := r.URL.Query().Get("page")
	pageInt, err := strconv.Atoi(pageStr)
	if err == nil || pageInt != 0 {
		page = pageInt
	}

	// Get category slug from URL
	slug := r.PathValue("category")

	// Generate template data
	// This is probably wasteful for non-existing category
	data := s.NewData(w, r)

	category, valid := isValidCategory(data.Categories, slug)
	if !valid {
		http.NotFound(w, r)
		return
	}

	// Pass category name as title of the page
	data.Title = category.Name

	var posts []database.Post

	// Use the generic cache wrapper
	err = redis.Cached(
		r.Context(),
		s.rdb,
		fmt.Sprintf("%s_posts_page_%d", slug, page),
		24*time.Hour,
		&posts,
		func() ([]database.Post, error) {
			return s.db.GetCategoryPosts(slug, page) // Call the actual underlying database method
		},
	)

	if err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	if len(posts) == 0 {
		log.Println("BINGOOOOOOOOOOOOO")
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

	data.Posts = posts
	if err := s.tm.Render(w, "category", data); err != nil {
		log.Println(err)
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
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

	dbStatus := s.db.Health()
	rdbStatus := s.rdb.Health(context.Background())

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
