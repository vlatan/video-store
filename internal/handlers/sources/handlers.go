package sources

import (
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
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
		"all:sources",
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

		s.tm.RenderHTML(w, r, "form", data)

		// // Fetch video data from YouTube
		// metadata, err := s.yt.GetVideos(videoID)
		// if err != nil {
		// 	formError.Message = utils.Capitalize(err.Error())
		// 	data.Form.Error = &formError
		// 	s.tm.RenderHTML(w, r, "form", data)
		// 	return
		// }

		// // Create post object
		// post := s.yt.CreatePost(metadata[0], "", "YouTube")
		// post.UserID = data.CurrentUser.ID

		// rowsAffected, err := s.sourcesRepo.InsertSource(r.Context(), post)
		// if err != nil || rowsAffected == 0 {
		// 	log.Printf("Could not insert the video '%s' in DB: %v", post.VideoID, err)
		// 	formError.Message = "Could not insert the video in DB"
		// 	data.Form.Error = &formError
		// 	s.tm.RenderHTML(w, r, "form", data)
		// 	return
		// }

		// redirectTo := fmt.Sprintf("/video/%s/", videoID)
		// http.Redirect(w, r, redirectTo, http.StatusFound)
	default:
		s.tm.HTMLError(w, r, http.StatusMethodNotAllowed, data)
	}
}
