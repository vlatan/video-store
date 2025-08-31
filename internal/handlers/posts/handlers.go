package posts

import (
	"errors"
	"factual-docs/internal/drivers/redis"
	"factual-docs/internal/models"
	"factual-docs/internal/utils"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

// Handle the Home page
func (s *Service) HomeHandler(w http.ResponseWriter, r *http.Request) {

	// Generate template data
	data := utils.GetDataFromContext(r)

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")
	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Construct the redis key
	redisKey := "posts"
	if orderBy == "likes" {
		redisKey += ":likes"
	}
	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	posts, err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		func() (*models.Posts, error) {
			return s.postsRepo.GetHomePosts(r.Context(), cursor, orderBy)
		},
	)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If there's a cursor this is not the first page, return JSON
	if cursor != "" {
		time.Sleep(time.Millisecond * 400)
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = posts
	s.ui.RenderHTML(w, r, "home.html", data)
}

// Handle posts in a certain category
func (s *Service) CategoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	slug := r.PathValue("category")
	cursor := r.URL.Query().Get("cursor")
	orderBy := r.URL.Query().Get("order_by")

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := utils.GetDataFromContext(r)

	// Construct the Redis key
	redisKey := fmt.Sprintf("category:%s:posts", slug)
	if orderBy == "likes" {
		redisKey += ":likes"
	}
	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	posts, err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		func() (*models.Posts, error) {
			return s.postsRepo.GetCategoryPosts(r.Context(), slug, cursor, orderBy)
		},
	)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If there's a cursor this is not the first page, return JSON
	if cursor != "" {
		time.Sleep(time.Millisecond * 400)
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = posts
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
	data := utils.GetDataFromContext(r)
	data.SearchQuery = searchQuery

	start := time.Now()
	encodedSearchQuery := utils.EscapeTrancateString(searchQuery, 100)

	// Construct the Redis key
	redisKey := fmt.Sprintf("posts:search:%s", encodedSearchQuery)
	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	posts, err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		func() (*models.Posts, error) {
			return s.postsRepo.SearchPosts(r.Context(), searchQuery, s.config.PostsPerPage, cursor)
		},
	)

	end := time.Since(start)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If there's a cursor this is not the first page, return JSON
	if cursor != "" {
		time.Sleep(time.Millisecond * 400)
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = posts
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())
	data.Title = "Search"
	s.ui.RenderHTML(w, r, "search.html", data)
}

// Handle adding new post via form
func (s *Service) NewPostHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := utils.GetDataFromContext(r)

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
		if s.postsRepo.PostExists(r.Context(), videoID) {
			formError.Message = "Video already posted"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Fetch video data from YouTube
		metadata, err := s.yt.GetVideos(r.Context(), videoID)
		if err != nil {
			log.Printf("Video '%s': %v", videoID, err)
			formError.Message = "Unable to fetch the video from YouTube"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Validate the video data
		if err := s.yt.ValidateYouTubeVideo(metadata[0]); err != nil {
			log.Printf("Video '%s': %v", videoID, err)
			formError.Message = utils.Capitalize(err.Error())
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Create post object
		post := s.yt.NewYouTubePost(metadata[0], "")
		post.UserID = data.CurrentUser.ID

		// Generate content using Gemini
		genaiResponse, err := s.gemini.GenerateInfo(
			r.Context(), post.Title, data.Categories,
		)

		if err != nil {
			log.Printf("Content generation using Gemini failed: %v", err)
		}

		post.Category = &models.Category{}
		if err == nil && genaiResponse != nil {
			post.ShortDesc = genaiResponse.Description
			post.Category.Name = genaiResponse.Category
		}

		// Insert the video
		rowsAffected, err := s.postsRepo.InsertPost(r.Context(), post)
		if err != nil || rowsAffected == 0 {
			log.Printf("Could not insert the video '%s' in DB: %v", post.VideoID, err)
			formError.Message = "Could not insert the video in DB"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

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

	// Generate the default data
	data := utils.GetDataFromContext(r)

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		http.NotFound(w, r)
		return
	}

	post, err := redis.GetItems(
		!data.CurrentUser.IsAuthenticated(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("post:%s", videoID),
		s.config.CacheTimeout,
		func() (*models.Post, error) {
			return s.postsRepo.GetSinglePost(r.Context(), videoID)
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}

	if err != nil {
		log.Printf("Error while getting the video '%s' from DB: %v", videoID, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// Assign the post to data
	data.CurrentPost = post
	data.Title = post.Title

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

	// Ignore the error on related posts, no posts will be shown
	relatedPosts, _ := redis.GetItems(
		!data.CurrentUser.IsAuthenticated(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("post:%s:related_posts", videoID),
		s.config.CacheTimeout,
		func() ([]models.Post, error) {
			return s.postsRepo.GetRelatedPosts(r.Context(), post.Title)
		},
	)

	data.CurrentPost.RelatedPosts = relatedPosts
	s.ui.RenderHTML(w, r, "post.html", data)
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
	allowedActions := []string{"like", "unlike", "fave", "unfave", "edit", "delete"}
	if !slices.Contains(allowedActions, action) {
		log.Printf("Not a valid action '%s' on video: %s\n", action, videoID)
		http.NotFound(w, r)
		return
	}

	// Get the current user
	user := utils.GetUserFromContext(r)

	// Check if user is authorized to edit or delete (admin)
	if (action == "edit" || action == "delete") &&
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
	case "edit":
		s.handleEdit(w, r, videoID, user)
	case "delete":
		s.handleBanPost(w, r, user.ID, videoID)
	default:
		utils.HttpError(w, http.StatusBadRequest)
	}
}
