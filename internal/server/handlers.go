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
	"fmt"
	"log"
	"net/http"
	"regexp"
	"slices"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/markbates/goth/gothic"
)

var validVideoID = regexp.MustCompile("^([-a-zA-Z0-9_]{11})$")

// Handle the Home page
func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {

	// Generate template data
	data := s.NewData(w, r)
	data.CurrentUser = s.getCurrentUser(w, r)

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
func (s *Server) categoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	// Get category slug from URL
	slug := r.PathValue("category")

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := s.NewData(w, r)
	data.CurrentUser = s.getCurrentUser(w, r)

	// Check if the category is valid
	category, valid := isValidCategory(data.Categories, slug)
	if !valid {
		log.Printf("Asked for invalid category '%s' on URI '%s'", slug, r.RequestURI)
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
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
	data.CurrentUser = s.getCurrentUser(w, r)
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
	data := s.NewData(w, r)
	data.CurrentUser = s.getCurrentUser(w, r)

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
	currentUser := s.getCurrentUser(w, r)

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
func (s *Server) handleEdit(w http.ResponseWriter, r *http.Request, videoID string, currentUser *templates.User) {
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

// Provider Auth
func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getRedirectPath(r)

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
// Wrap this with middleware to allow only authnenticated users
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {

	// The origin URL of the user
	redirectTo := getRedirectPath(r)

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

// Delete the user account
// Wrap this with middleware to allow only authnenticated users
func (s *Server) deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	// This is a POST request, close the body
	defer r.Body.Close()

	// The origin URL of the user
	redirectTo := getRedirectPath(r)

	// Get the current user
	currentUser := s.getCurrentUser(w, r)

	// Remove gothic session if any
	if err := gothic.Logout(w, r); err != nil {
		log.Printf("Error loging out the user with gothic: %v", err)
		s.storeFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Remove user session
	if err := s.logoutUser(w, r); err != nil {
		log.Printf("Error loging out the user: %v", err)
		s.storeFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Delete the user from DB
	rowsAffected, err := s.db.Exec(r.Context(), database.DeleteUserQuery, currentUser.ID)
	if err != nil {
		log.Printf("Could not delete user %d: %v", currentUser.ID, err)
		s.storeFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	if rowsAffected == 0 {
		log.Printf("No such user %d to delete", currentUser.ID)
		s.storeFlashMessage(w, r, &failedDeleteAccount)
		http.Redirect(w, r, redirectTo, http.StatusFound)
		return
	}

	// Attempt to remove the avatar from disk and redis
	s.deleteAvatar(r, currentUser.AnalyticsID)

	// Attempt to send revoke request
	if currentUser.AccessToken != "" {
		revokeLogin(currentUser)
	}

	s.storeFlashMessage(w, r, &successDeleteAccount)
	http.Redirect(w, r, redirectTo, http.StatusFound)

}

// Send revoke request. It will work if the access token is not expired.
func revokeLogin(user *templates.User) (response *http.Response, err error) {

	switch user.Provider {
	case "google":
		url := "https://oauth2.googleapis.com/revoke"
		contentType := "application/x-www-form-urlencoded"
		body := []byte("token=" + user.AccessToken)
		response, err = http.Post(url, contentType, bytes.NewBuffer(body))
	case "facebook":
		url := fmt.Sprintf("https://graph.facebook.com/v23.0/%s/permissions", user.UserID)
		body := []byte("access_token=" + user.AccessToken)
		req, reqErr := http.NewRequest("DELETE", url, bytes.NewBuffer(body))
		if reqErr != nil {
			return response, reqErr
		}
		client := &http.Client{}
		response, err = client.Do(req)
	}

	if err != nil {
		return response, err
	}

	defer response.Body.Close()
	return response, err
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
// Defaults to 1 if invalid page
func getPageNum(r *http.Request) (page int) {
	pageStr := r.URL.Query().Get("page")
	if pageInt, err := strconv.Atoi(pageStr); err == nil {
		page = pageInt
	}

	// Do not allow negative or zero pages
	if page <= 0 {
		page = 1
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
