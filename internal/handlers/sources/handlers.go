package sources

import (
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/utils"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Handle all sources page
func (s *Service) SourcesHandler(w http.ResponseWriter, r *http.Request) {
	// Generate template data
	data := s.ui.NewData(w, r)

	// Get sources from redis or DB
	var sources []models.Source
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		"sources",
		s.config.CacheTimeout,
		&sources,
		func() ([]models.Source, error) {
			return s.sourcesRepo.GetSources(r.Context())
		},
	)

	if err != nil {
		log.Printf("Was unabale to fetch sources on URI '%s': %v", r.RequestURI, err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if len(sources) == 0 {
		log.Printf("Fetched zero sources on URI '%s'", r.RequestURI)
		s.ui.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	data.Sources = sources
	data.Title = "Sources"
	s.ui.RenderHTML(w, r, "sources.html", data)

}

// Handle adding new post via form
func (s *Service) NewSourceHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := s.ui.NewData(w, r)

	// Populate needed data for an empty form
	data.Form = &models.Form{
		Legend: "New Playlist",
		Content: &models.FormGroup{
			Label:       "Post YouTube Playlist URL",
			Placeholder: "Playlist URL here...",
		},
	}
	data.Title = "Add New Source"

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
		playlistID, err := extractPlaylistID(url)
		if err != nil {
			formError.Message = "Could not extract the playlist ID"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check if the playlist is already posted
		if s.sourcesRepo.SourceExists(r.Context(), playlistID) {
			formError.Message = "Playlist already posted"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Fetch playlist metadata from YouTube
		sources, err := s.yt.GetSources(playlistID)
		if err != nil {
			log.Printf("Playlist '%s': %v", playlistID, err)
			formError.Message = "Unable to fetch the playlist from YouTube"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Fetch channel data from YouTube
		channelID := sources[0].Snippet.ChannelId
		channels, err := s.yt.GetChannels(channelID)
		if err != nil {
			log.Printf("Channel '%s': %v", channelID, err)
			formError.Message = "Unable to fetch channel info from YouTube"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Create a source object
		source := s.yt.NewYouTubeSource(sources[0], channels[0])
		source.UserID = data.CurrentUser.ID

		// Insert the source in DB
		rowsAffected, err := s.sourcesRepo.InsertSource(r.Context(), source)
		if err != nil || rowsAffected == 0 {
			log.Printf("Could not insert the source '%s' in DB: %v", source.PlaylistID, err)
			formError.Message = "Could not insert the source in DB"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the souurce
		redirectTo := fmt.Sprintf("/source/%s/", playlistID)
		http.Redirect(w, r, redirectTo, http.StatusFound)

	default:
		s.ui.HTMLError(w, r, http.StatusMethodNotAllowed, data)
	}
}

// Handle posts in a certain source
func (s *Service) SourcePostsHandler(w http.ResponseWriter, r *http.Request) {

	// Get category slug from URL
	sourceID := r.PathValue("source")

	// Generate template data (it gets all the categories too)
	// This is probably wasteful for non-existing category
	data := s.ui.NewData(w, r)

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
			s.ui.JSONError(w, r, http.StatusInternalServerError)
			return
		}
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if len(posts.Items) == 0 {
		log.Printf("Fetched zero posts on URI '%s'", r.RequestURI)
		if page > 1 {
			s.ui.JSONError(w, r, http.StatusNotFound)
			return
		}
		s.ui.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	// If not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		s.ui.WriteJSON(w, r, posts.Items)
		return
	}

	data.Posts = posts
	if sourceID == "other" {
		data.Posts.Title = "Other Uploads"
	}
	data.Title = data.Posts.Title
	s.ui.RenderHTML(w, r, "source.html", data)
}
