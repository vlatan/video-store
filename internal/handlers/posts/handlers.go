package posts

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/handlers/auth"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/redirect"
	"github.com/vlatan/video-store/internal/utils"

	"github.com/jackc/pgx/v5"
)

const postCacheKey = "post:%s"
const relatedPostsCacheKey = "post:%s:related_posts"

// Handle the Home page
func (s *Service) HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Construct the redis key
	redisKey := "home:posts"
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	// Generate template data
	data := models.GetDataFromContext(r)

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the home results only for the admin
	if data.IsCurrentUserAdmin() {
		posts, err = s.postsRepo.GetHomePosts(
			r.Context(), "", orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetHomePosts(
					r.Context(), "", orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get posts from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	data.Posts = &posts
	s.ui.RenderHTML(w, r, "home.html", data)
}

// Handle posts in a certain category
func (s *Service) CategoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("category")
	orderBy := r.URL.Query().Get("order_by")

	// Construct the Redis key
	redisKey := fmt.Sprintf("category:%s:posts", slug)
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := models.GetDataFromContext(r)

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the category posts only for the admin
	if data.IsCurrentUserAdmin() {
		posts, err = s.postsRepo.GetCategoryPosts(
			r.Context(), slug, "", orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetCategoryPosts(
					r.Context(), slug, "", orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed get posts from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		http.NotFound(w, r)
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

	// Generate the default data
	data := models.GetDataFromContext(r)
	data.SearchQuery = searchQuery

	start := time.Now()
	encodedSearchQuery := utils.EscapeTrancateString(searchQuery, 100)

	// Construct the Redis key
	redisKey := fmt.Sprintf("posts:search:%s", encodedSearchQuery)

	var (
		err   error
		posts models.Posts
	)

	// Don't cache the search results only for the admin
	if data.IsCurrentUserAdmin() {
		posts, err = s.postsRepo.SearchPosts(
			r.Context(), searchQuery, s.config.PostsPerPage, "",
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.SearchPosts(
					r.Context(), searchQuery, s.config.PostsPerPage, "",
				)
			},
		)
	}

	end := time.Since(start)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed get posts from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
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
			slog.ErrorContext(
				r.Context(), "failed get video data from YouTube",
				"path", r.URL.Path,
				"error", err,
			)
			formError.Message = "Unable to fetch the video from YouTube"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Validate the video data
		if err := s.yt.ValidateYouTubeVideo(metadata[0]); err != nil {
			slog.ErrorContext(
				r.Context(), "failed get validate this video",
				"path", r.URL.Path,
				"error", err,
			)
			formError.Message = utils.Capitalize(err.Error())
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Create post object
		post := s.yt.NewYouTubePost(metadata[0], "")
		post.UserActions = &models.Actions{UserID: data.CurrentUser.ID}

		// Insert the video
		rowsAffected, err := s.postsRepo.InsertPost(r.Context(), post)
		if err != nil || rowsAffected == 0 {
			slog.ErrorContext(
				r.Context(), "failed to insert the post in DB",
				"path", r.URL.Path,
				"error", err,
			)
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
		redirectURL := fmt.Sprintf("/video/%s/", videoID)
		redirectTo := redirect.Sanitize(redirectURL, auth.IsProtectedRoute)
		redirect.Execute(w, r, redirectTo, http.StatusFound)

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
		slog.ErrorContext(
			r.Context(), "failed to get the post from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Assign the post to data
	data.CurrentPost = &post

	// Check whether the current user liked and/or faved the post
	if data.CurrentUser.IsAuthenticated() {
		userActions, _ := s.usersRepo.GetUserActions(
			r.Context(),
			data.CurrentUser.ID,
			data.CurrentPost.ID,
		)
		data.CurrentPost.UserActions = &userActions
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
		slog.ErrorContext(
			r.Context(), "failed to get the post from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Generate default data
	data := models.GetDataFromContext(r)

	// Assign post data
	data.CurrentPost = &post
	if data.CurrentPost.Category == nil {
		data.CurrentPost.Category = &models.Category{}
	}

	// Populate needed data for the post form
	data.Form = &models.Form{
		Legend: "Edit Post",
		Title: &models.FormGroup{
			Label:       "Title",
			Placeholder: "Your title...",
			Value:       data.CurrentPost.GetTitle(),
		},
		Content: &models.FormGroup{
			Type:        models.FieldTypeTextarea,
			Label:       "Content",
			Placeholder: "You can use markdown...",
			Value:       data.CurrentPost.Summary,
		},
		Category: &models.FormGroup{
			Label: "Category",
			Value: data.CurrentPost.Category.Name,
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
			data.Form.Category.Value, // category name
			data.Form.Content.Value,  // summary
		)

		if err != nil || rowsAffected == 0 {
			slog.ErrorContext(
				r.Context(), "failed to update the post in DB",
				"path", r.URL.Path,
				"error", err,
			)
			formError.Message = "Could not update the post in DB"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Delete the redis cache
		redisKey := fmt.Sprintf(postCacheKey, videoID)
		if err = s.rdb.Client.Del(r.Context(), redisKey).Err(); err != nil {
			slog.ErrorContext(
				r.Context(), "failed to delete the cache on post",
				"path", r.URL.Path,
				"error", err,
			)
			formError.Message = "Could not delete the cache on post"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the updated page
		redirectURL := fmt.Sprintf("/video/%s/", videoID)
		redirectTo := redirect.Sanitize(redirectURL, auth.IsProtectedRoute)
		redirect.Execute(w, r, redirectTo, http.StatusFound)

	default:
		utils.HttpError(w, http.StatusMethodNotAllowed)
	}
}

// Handle a post ban
func (s *Service) BanPostHandler(w http.ResponseWriter, r *http.Request) {

	// Validate the YT ID
	videoID := r.PathValue("video")
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
		return
	}

	// Get the current user
	user := models.GetUserFromContext(r)

	rowsAffected, err := s.postsRepo.BanPost(r.Context(), videoID)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "user failed to ban/delete the video",
			"path", r.URL.Path,
			"userId", user.ID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.NotFound(w, r)
		return
	}

	successDelete := models.FlashMessage{
		Message:  "The video has been deleted!",
		Category: "info",
	}

	s.ui.StoreFlashMessage(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
