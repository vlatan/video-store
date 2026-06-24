package posts

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"github.com/jackc/pgx/v5"
)

const postCacheKey = "post:%s"
const relatedPostsCacheKey = "post:%s:related_posts"

// Handle the Home page
func (s *Service) HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Generate template data
	data := models.GetDataFromContext(r)

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")
	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Construct the redis key
	redisKey := "home:posts"
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the home results only for the admin
	if data.IsCurrentUserAdmin() {
		posts, err = s.postsRepo.GetHomePosts(
			r.Context(), cursor, orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetHomePosts(
					r.Context(), cursor, orderBy,
				)
			},
		)
	}

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI %q: %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If there's a cursor this is not the first page, return JSON
	if cursor != "" {
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = &posts
	s.ui.RenderHTML(w, r, "home.html", data)
}

// Handle posts in a certain category
func (s *Service) CategoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("category")
	cursor := r.URL.Query().Get("cursor")
	orderBy := r.URL.Query().Get("order_by")

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := models.GetDataFromContext(r)

	// Construct the Redis key
	redisKey := fmt.Sprintf("category:%s:posts", slug)
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the category posts only for the admin
	if data.IsCurrentUserAdmin() {
		posts, err = s.postsRepo.GetCategoryPosts(
			r.Context(), slug, cursor, orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetCategoryPosts(
					r.Context(), slug, cursor, orderBy,
				)
			},
		)
	}

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI %q: %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If there's a cursor this is not the first page, return JSON
	if cursor != "" {
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = &posts
	data.Title = data.Posts.Title
	s.ui.RenderHTML(w, r, "category.html", data)
}

// Handle the requests from the searchform
func (s *Service) SearchPostsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the search query
	searchQuery := r.URL.Query().Get("q")
	if searchQuery == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get the cursor if any
	cursor := r.URL.Query().Get("cursor")

	// Generate the default data
	data := models.GetDataFromContext(r)
	data.SearchQuery = searchQuery

	start := time.Now()
	encodedSearchQuery := utils.EscapeTrancateString(searchQuery, 100)

	// Construct the Redis key
	redisKey := fmt.Sprintf("posts:search:%s", encodedSearchQuery)
	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the search results only for the admin
	if data.IsCurrentUserAdmin() {
		posts, err = s.postsRepo.SearchPosts(
			r.Context(), searchQuery, s.config.PostsPerPage, cursor,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.SearchPosts(
					r.Context(), searchQuery, s.config.PostsPerPage, cursor,
				)
			},
		)
	}

	end := time.Since(start)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI %q: %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If there's a cursor this is not the first page, return JSON
	if cursor != "" {
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = &posts
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())
	data.Title = "Search"
	s.ui.RenderHTML(w, r, "search.html", data)
}

// Handle adding new post via form
func (s *Service) NewPostHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := models.GetDataFromContext(r)

	// Populate needed data for an empty form
	data.Form = &models.Form{
		Legend: "New Video",
		Content: &models.FormGroup{
			Label:       "Post YouTube Video URL",
			Placeholder: "Video URL here...",
		},
	}
	data.Title = "Add New Video"

	switch r.Method {
	case "GET":
		// Serve the page with the form
		s.ui.RenderHTML(w, r, "form.html", data)

	case "POST":

		var formError models.FlashMessage

		err := r.ParseForm()
		if err != nil {
			formError.Message = "Could not parse the form"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Get the URL from the form
		url := r.FormValue("content")
		data.Form.Content.Value = url

		// Exctract the ID from the URL
		videoID, err := extractYouTubeID(url)
		if err != nil {
			formError.Message = "Could not extract the video ID"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Validate the YT ID
		if validVideoID.FindStringSubmatch(videoID) == nil {
			formError.Message = "Could not validate the video ID"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check if the video is already posted
		if err = s.postsRepo.PostExists(r.Context(), videoID); err == nil {
			formError.Message = "Video already posted"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Fetch video data from YouTube
		metadata, err := s.yt.GetVideos(
			r.Context(),
			&utils.RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Second,
				Delay:      time.Second,
			},
			videoID,
		)

		if err != nil {
			log.Printf("Video %q: %v", videoID, err)
			formError.Message = "Unable to fetch the video from YouTube"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Validate the video data
		if err := s.yt.ValidateYouTubeVideo(metadata[0]); err != nil {
			log.Printf("Video %q: %v", videoID, err)
			formError.Message = utils.Capitalize(err.Error())
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Create post object
		post := s.yt.NewYouTubePost(metadata[0], "")
		post.UserID = data.CurrentUser.ID

		// Insert the video
		rowsAffected, err := s.postsRepo.InsertPost(r.Context(), post)
		if err != nil || rowsAffected == 0 {
			log.Printf("Could not insert the video %q in DB: %v", post.VideoID, err)
			formError.Message = "Could not insert the video in DB"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Generate content in the background using Gemini.
		// Give reasonable TTL for this to finish.
		// In production no need to use it, the worker will
		// generate the post content overnight.
		go func() {
			if err := s.generatePostContent(r, post, 30*time.Minute); err != nil {
				log.Println(err)
			}
		}()

		// Check out the video
		redirectTo := fmt.Sprintf("/video/%s/", videoID)
		http.Redirect(w, r, redirectTo, http.StatusFound)

	default:
		utils.HttpError(w, http.StatusMethodNotAllowed)
	}
}

// Handle a single post
func (s *Service) SinglePostHandler(w http.ResponseWriter, r *http.Request) {

	// Get video id from URL path
	videoID := r.PathValue("video")

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
		return
	}

	// Generate the default data
	data := models.GetDataFromContext(r)

	var (
		err  error
		post models.Post
	)

	// Don't cache single post for logged in users
	if data.CurrentUser.IsAuthenticated() {
		post, err = s.postsRepo.GetSinglePost(r.Context(), videoID)
	} else {
		post, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			fmt.Sprintf(postCacheKey, videoID),
			s.config.CacheTimeout,
			func() (models.Post, error) {
				return s.postsRepo.GetSinglePost(r.Context(), videoID)
			},
		)
	}

	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf("Error while getting the video %q from DB: %v", videoID, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Assign the post to data
	data.CurrentPost = &post

	// Check whether the current user liked and/or faved the post
	if data.CurrentUser.IsAuthenticated() {
		userActions, _ := s.postsRepo.GetUserActions(
			r.Context(),
			data.CurrentUser.ID,
			data.CurrentPost.ID,
		)
		data.CurrentPost.UserLiked = userActions.Liked
		data.CurrentPost.UserFaved = userActions.Faved
	}

	// Don't cache the related posts only for the admin.
	// Ignore the error on related posts, no posts will be shown.
	var relatedPosts models.Posts
	if data.IsCurrentUserAdmin() {
		relatedPosts, _ = s.postsRepo.GetRelatedPosts(r.Context(), post.GetTitle())
	} else {
		relatedPosts, _ = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			fmt.Sprintf(relatedPostsCacheKey, videoID),
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetRelatedPosts(r.Context(), post.GetTitle())
			},
		)
	}

	data.CurrentPost.RelatedPosts = relatedPosts.Items
	s.ui.RenderHTML(w, r, "post.html", data)
}

// UpdatePostHandler handles the post update
func (s *Service) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {

	// Get video id from URL path
	videoID := r.PathValue("video")

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
		return
	}

	// Get the post data straight from DB
	post, err := s.postsRepo.GetSinglePost(r.Context(), videoID)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf("Error while getting the post %q from DB: %v", videoID, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Generate default data
	data := models.GetDataFromContext(r)

	// Assign page data
	data.CurrentPost = &post
	if data.CurrentPost.Category == nil {
		data.CurrentPost.Category = &models.Category{}
	}

	// Populate needed data for the page form
	data.Form = &models.Form{
		Legend: "Edit Post",
		Title: &models.FormGroup{
			Label:       "Title",
			Placeholder: "Your title...",
			Value:       data.CurrentPost.Title,
		},
		Content: &models.FormGroup{
			Type:        models.FieldTypeTextarea,
			Label:       "Content",
			Placeholder: "You can use markdown...",
			Value:       data.CurrentPost.Summary,
		},
		Category: &models.FormGroup{
			Label: "Category",
			Value: data.CurrentPost.Category.Slug,
		},
	}

	data.Title = "Edit This Post"

	switch r.Method {
	case "GET":
		// Serve the page with the form
		s.ui.RenderHTML(w, r, "form.html", data)

	case "POST":
		var formError models.FlashMessage

		err := r.ParseForm()
		if err != nil {
			formError.Message = "Could not parse the form"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Get the title and the content from the form
		data.Form.Title.Value = r.FormValue("title")
		data.Form.Category.Value = r.FormValue("category")
		data.Form.Content.Value = r.FormValue("content")

		// Update the page
		rowsAffected, err := s.postsRepo.UpdatePost(
			r.Context(),
			videoID,
			data.Form.Title.Value,    // original title
			data.Form.Category.Value, // category slug
			data.Form.Content.Value,  // summary
		)

		if err != nil || rowsAffected == 0 {
			log.Printf("Could not update the post %q in DB: %v", videoID, err)
			formError.Message = "Could not update the post in DB"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Delete the redis cache
		redisKey := fmt.Sprintf(postCacheKey, videoID)
		if err = s.rdb.Client.Del(r.Context(), redisKey).Err(); err != nil {
			log.Printf("could not delete the cache on post %q; %v", videoID, err)
			formError.Message = "Could not delete the cache on post"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the updated page
		redirectTo := fmt.Sprintf("/video/%s/", videoID)
		http.Redirect(w, r, redirectTo, http.StatusFound)

	default:
		utils.HttpError(w, http.StatusMethodNotAllowed)
	}
}

// Perform an action on a video
func (s *Service) PostActionHandler(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
		return
	}

	// Validate the action
	action := r.PathValue("action")
	allowedActions := []string{"like", "unlike", "fave", "unfave", "delete"}
	if !slices.Contains(allowedActions, action) {
		log.Printf("Not a valid action %q on video: %s\n", action, videoID)
		http.NotFound(w, r)
		return
	}

	// Get the current user
	user := models.GetUserFromContext(r)

	// Check if user is authorized to edit or delete (admin)
	if action == "delete" &&
		!user.IsAdmin(s.config.AdminProviderUserId, s.config.AdminProvider) {
		utils.HttpError(w, http.StatusForbidden)
		return
	}

	switch action {
	case "like":
		s.handleLike(w, r, user.ID, videoID)
	case "unlike":
		s.handleUnlike(w, r, user.ID, videoID)
	case "fave":
		s.handleFave(w, r, user.ID, videoID)
	case "unfave":
		s.handleUnfave(w, r, user.ID, videoID)
	case "delete":
		s.handleBanPost(w, r, user.ID, videoID)
	default:
		utils.HttpError(w, http.StatusBadRequest)
	}
}
