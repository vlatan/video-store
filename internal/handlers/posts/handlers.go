package posts

import (
	"errors"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/utils"
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
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

	// Get page number from a query param
	page := utils.GetPageNum(r)
	redisKey := fmt.Sprintf("posts:page:%d", page)

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	var posts []models.Post
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		&posts,
		func() ([]models.Post, error) {
			return s.postsRepo.GetPosts(r.Context(), page, orderBy)
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

	data.Posts = &models.Posts{}
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
	data.CurrentUser = s.auth.GetUserFromContext(r)

	// Get page number from a query param
	page := utils.GetPageNum(r)
	redisKey := fmt.Sprintf("category:%s:posts:page:%d", slug, page)

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	var posts = &models.Posts{}
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		&posts,
		func() (*models.Posts, error) {
			return s.postsRepo.GetCategoryPosts(r.Context(), slug, orderBy, page)
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

	if len(posts.Items) == 0 {
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
		s.tm.WriteJSON(w, r, posts.Items)
		return
	}

	data.Posts = posts
	data.Title = data.Posts.Title
	s.tm.RenderHTML(w, r, "category", data)
}

// Handle posts in a certain source
func (s *Service) SourcePostsHandler(w http.ResponseWriter, r *http.Request) {

	// Get category slug from URL
	sourceID := r.PathValue("source")

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

	// Get page number from a query param
	page := utils.GetPageNum(r)
	redisKey := fmt.Sprintf("source:%s:posts:page:%d", sourceID, page)

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	var posts = &models.Posts{}
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		&posts,
		func() (*models.Posts, error) {
			return s.postsRepo.GetSourcePosts(r.Context(), sourceID, orderBy, page)
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

	if len(posts.Items) == 0 {
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
		s.tm.WriteJSON(w, r, posts.Items)
		return
	}

	data.Posts = posts
	data.Title = data.Posts.Title
	s.tm.RenderHTML(w, r, "source", data)
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
	data.CurrentUser = s.auth.GetUserFromContext(r)
	data.SearchQuery = searchQuery

	limit := s.config.PostsPerPage
	offset := (page - 1) * limit

	start := time.Now()
	encodedSearchQuery := utils.EscapeTrancateString(searchQuery, 100)

	// For search posts we are using the database.Posts struct,
	// so we can add total results and time took
	var posts models.Posts
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("posts:search:page:%d:%s", page, encodedSearchQuery),
		s.config.CacheTimeout,
		&posts,
		func() (models.Posts, error) {
			return s.postsRepo.SearchPosts(r.Context(), searchQuery, limit, offset)
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

	data.Posts = &posts
	data.Posts.TimeTook = fmt.Sprintf("%.2f", end.Seconds())
	data.Title = "Search"
	s.tm.RenderHTML(w, r, "search", data)
}

// Handle adding new post via form
func (s *Service) NewPostHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

	// Populate needed data for an empty form
	data.Form = &models.Form{}
	data.Form.Legend = "New Video"
	data.Form.Content.Label = "Post YouTube Video URL"
	data.Form.Content.Placeholder = "Video URL here..."
	data.Title = "Add New Video"

	switch r.Method {
	case "GET":
		// Serve the page with the form
		s.tm.RenderHTML(w, r, "form", data)

	case "POST":

		var formError models.FlashMessage

		err := r.ParseForm()
		if err != nil {
			formError.Message = "Could not parse the form"
			data.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
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
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Validate the YT ID
		if validVideoID.FindStringSubmatch(videoID) == nil {
			formError.Message = "Could not validate the video ID"
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Check if the video is already posted
		if s.postsRepo.PostExists(r.Context(), videoID) {
			formError.Message = "Video already posted"
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Fetch video data from YouTube
		metadata, err := s.yt.GetVideos(videoID)
		if err != nil {
			formError.Message = utils.Capitalize(err.Error())
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Validate the video data
		if err := s.yt.ValidateYouTubeVideo(metadata[0]); err != nil {
			formError.Message = utils.Capitalize(err.Error())
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Create post object
		post := s.yt.NewYouTubePost(metadata[0], "")
		post.UserID = data.CurrentUser.ID

		// Generate content using Gemini
		gc, err := s.gemini.GenerateInfo(r.Context(), post.Title, data.Categories)
		if err != nil {
			log.Printf("Content generation using Gemini failed: %v", err)
		}

		if gc != nil {
			post.ShortDesc = gc.Description
			post.Category = &models.Category{Name: gc.Category}
		}

		// Insert the video
		rowsAffected, err := s.postsRepo.InsertPost(r.Context(), post)
		if err != nil || rowsAffected == 0 {
			log.Printf("Could not insert the video '%s' in DB: %v", post.VideoID, err)
			formError.Message = "Could not insert the video in DB"
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Check out the video
		redirectTo := fmt.Sprintf("/video/%s/", videoID)
		http.Redirect(w, r, redirectTo, http.StatusFound)

	default:
		s.tm.HTMLError(w, r, http.StatusMethodNotAllowed, data)
	}
}

// Handle a single post
func (s *Service) SinglePostHandler(w http.ResponseWriter, r *http.Request) {
	// Get video id from URL path
	videoID := r.PathValue("video")

	// Generate the default data
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		log.Println("Not a valid video ID:", videoID)
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	var post models.Post
	err := redis.GetItems(
		!data.CurrentUser.IsAuthenticated(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("post:%s", videoID),
		s.config.CacheTimeout,
		&post,
		func() (models.Post, error) {
			return s.postsRepo.GetSinglePost(r.Context(), videoID)
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
		userActions, _ := s.postsRepo.GetUserActions(
			r.Context(),
			data.CurrentUser.ID,
			data.CurrentPost.ID,
		)
		data.CurrentPost.CurrentUserLiked = userActions.Liked
		data.CurrentPost.CurrentUserFaved = userActions.Faved
	}

	// Ignore the error on related posts, no posts will be shown
	var relatedPosts []models.Post
	redis.GetItems(
		!data.CurrentUser.IsAuthenticated(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("post:%s:related_posts", videoID),
		s.config.CacheTimeout,
		&relatedPosts,
		func() ([]models.Post, error) {
			return s.postsRepo.GetRelatedPosts(r.Context(), post.Title)
		},
	)

	data.CurrentPost.RelatedPosts = relatedPosts
	s.tm.RenderHTML(w, r, "post", data)
}

// Perform an action on a video
func (s *Service) PostActionHandler(w http.ResponseWriter, r *http.Request) {

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
	currentUser := s.auth.GetUserFromContext(r)

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
