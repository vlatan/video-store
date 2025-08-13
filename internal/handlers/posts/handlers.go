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

	// Get page number from a query param
	page := utils.GetPageNum(r)
	redisKey := fmt.Sprintf("posts:page:%d", page)

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	posts, err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		func() ([]models.Post, error) {
			return s.postsRepo.GetHomePosts(r.Context(), page, orderBy)
		},
	)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if len(posts) == 0 {
		http.NotFound(w, r)
		return
	}

	// If not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = &models.Posts{}
	data.Posts.Items = posts
	s.ui.RenderHTML(w, r, "home.html", data)
}

// Handle posts in a certain category
func (s *Service) CategoryPostsHandler(w http.ResponseWriter, r *http.Request) {

	// Get category slug from URL
	slug := r.PathValue("category")

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := utils.GetDataFromContext(r)

	// Get page number from a query param
	page := utils.GetPageNum(r)
	redisKey := fmt.Sprintf("category:%s:posts:page:%d", slug, page)

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")
	if orderBy == "likes" {
		redisKey += ":likes"
	}

	posts, err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		redisKey,
		s.config.CacheTimeout,
		func() (*models.Posts, error) {
			return s.postsRepo.GetCategoryPosts(r.Context(), slug, orderBy, page)
		},
	)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		http.NotFound(w, r)
		return
	}

	// if not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		s.ui.WriteJSON(w, r, posts.Items)
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

	// Get the page number from the request query param
	page := utils.GetPageNum(r)

	// Generate the default data
	data := utils.GetDataFromContext(r)
	data.SearchQuery = searchQuery

	limit := s.config.PostsPerPage
	offset := (page - 1) * limit

	start := time.Now()
	encodedSearchQuery := utils.EscapeTrancateString(searchQuery, 100)

	// For search posts we are using the database.Posts struct,
	// so we can add total results and time took
	posts, err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("posts:search:page:%d:%s", page, encodedSearchQuery),
		s.config.CacheTimeout,
		func() (*models.Posts, error) {
			return s.postsRepo.SearchPosts(r.Context(), searchQuery, limit, offset)
		},
	)

	end := time.Since(start)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		s.ui.WriteJSON(w, r, posts.Items)
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
		log.Println("Not a valid video ID:", videoID)
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
		data.CurrentPost.CurrentUserLiked = userActions.Liked
		data.CurrentPost.CurrentUserFaved = userActions.Faved
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

	// Get the current user
	currentUser := utils.GetUserFromContext(r)

	// Check if user is authorized to edit or delete (admin)
	if (action == "edit" || action == "delete") &&
		currentUser.UserID != s.config.AdminOpenID {
		utils.HttpError(w, http.StatusForbidden)
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
		s.handleBanPost(w, r, currentUser.ID, videoID)
	default:
		utils.HttpError(w, http.StatusBadRequest)
	}
}
