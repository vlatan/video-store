package sources

import (
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/utils"
	"fmt"
	"log"
	"net/http"
)

// Handle all sources page
func (s *Service) SourcesHandler(w http.ResponseWriter, r *http.Request) {
	// Generate template data
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

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
		s.tm.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if len(sources) == 0 {
		log.Printf("Fetched zero sources on URI '%s'", r.RequestURI)
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	data.Sources = sources
	data.Title = "Sources"
	s.tm.RenderHTML(w, r, "sources", data)

}

// Handle adding new post via form
func (s *Service) NewSourceHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

	// Populate needed data for an empty form
	data.Form = &models.Form{}
	data.Form.Legend = "New Playlist"
	data.Form.Content.Label = "Post YouTube Playlist URL"
	data.Form.Content.Placeholder = "Playlist URL here..."

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
		playlistID, err := extractPlaylistID(url)
		if err != nil {
			formError.Message = "Could not extract the playlist ID"
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Check if the playlist is already posted
		if s.sourcesRepo.SourceExists(r.Context(), playlistID) {
			formError.Message = "Source already posted"
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Fetch playlist metadata from YouTube
		sources, err := s.yt.GetSources(playlistID)
		if err != nil {
			formError.Message = utils.Capitalize(err.Error())
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Fetch channel data from YouTube
		channels, err := s.yt.GetChannels(sources[0].Snippet.ChannelId)
		if err != nil {
			formError.Message = utils.Capitalize(err.Error())
			data.Form.Error = &formError
			s.tm.RenderHTML(w, r, "form", data)
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
			s.tm.RenderHTML(w, r, "form", data)
			return
		}

		// Check out the souurce
		redirectTo := fmt.Sprintf("/source/%s/", playlistID)
		http.Redirect(w, r, redirectTo, http.StatusFound)

	default:
		s.tm.HTMLError(w, r, http.StatusMethodNotAllowed, data)
	}
}
