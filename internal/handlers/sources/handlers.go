package sources

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Handle all sources page
func (s *Service) SourcesHandler(w http.ResponseWriter, r *http.Request) {

	// Generate template data
	data := models.GetDataFromContext(r)

	var (
		err     error
		sources models.Sources
	)

	if data.IsCurrentUserAdmin() {
		sources, err = s.sourcesRepo.GetSources(r.Context())
	} else {
		sources, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			"sources",
			s.config.CacheTimeout,
			func() (models.Sources, error) {
				return s.sourcesRepo.GetSources(r.Context())
			},
		)
	}

	if err != nil {
		log.Printf("Was unabale to fetch sources on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if len(sources) == 0 {
		http.NotFound(w, r)
		return
	}

	data.Sources = sources
	data.Title = "Sources"
	s.ui.RenderHTML(w, r, "sources.html", data)

}

// Handle adding new post via form
func (s *Service) NewSourceHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := models.GetDataFromContext(r)

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
		sources, err := s.yt.GetSources(
			r.Context(),
			&utils.RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Second,
				Delay:      time.Second,
			},
			playlistID,
		)
		if err != nil {
			log.Printf("Playlist '%s': %v", playlistID, err)
			formError.Message = "Unable to fetch the playlist from YouTube"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Fetch channel data from YouTube
		channelID := sources[0].Snippet.ChannelId
		channels, err := s.yt.GetChannels(
			r.Context(),
			&utils.RetryConfig{
				MaxRetries: 3,
				MaxJitter:  time.Second,
				Delay:      time.Second,
			},
			channelID,
		)
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
		utils.HttpError(w, http.StatusMethodNotAllowed)
	}
}

// Handle posts in a certain source
func (s *Service) SourcePostsHandler(w http.ResponseWriter, r *http.Request) {

	sourceID := r.PathValue("source")
	cursor := r.URL.Query().Get("cursor")
	orderBy := r.URL.Query().Get("order_by")

	// Generate template data
	data := models.GetDataFromContext(r)

	// Construct the Redis key
	redisKey := fmt.Sprintf("source:%s:posts", sourceID)
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

	if data.IsCurrentUserAdmin() {
		posts, err = s.postsRepo.GetSourcePosts(
			r.Context(), sourceID, cursor, orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (models.Posts, error) {
				return s.postsRepo.GetSourcePosts(
					r.Context(), sourceID, cursor, orderBy,
				)
			},
		)
	}

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If there's a cursor this is not the first page, return JSON
	if cursor != "" {
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = &posts
	if sourceID == "other" {
		data.Posts.Title = "Other Uploads"
	}
	data.Title = data.Posts.Title
	s.ui.RenderHTML(w, r, "source.html", data)
}
