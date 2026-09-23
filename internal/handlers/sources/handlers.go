package sources

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/handlers/auth"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/ctxv"
	"github.com/vlatan/video-store/internal/utils/redirect"
	"github.com/vlatan/video-store/internal/utils/retry"
)

// Handle all sources page
func (s *Service) SourcesHandler(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := ctxv.Get[*types.TemplateData](r.Context())

	var (
		err     error
		sources types.Sources
	)

	if data.CurrentUser.IsAdmin() {
		sources, err = s.sourcesRepo.GetAllSources(r.Context())
	} else {
		sources, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			"sources",
			s.config.CacheTimeout,
			func() (types.Sources, error) {
				return s.sourcesRepo.GetAllSources(r.Context())
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get sources from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	if len(sources) == 0 {
		slog.WarnContext(r.Context(), "no sources found in DB")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	data.Sources = sources
	data.Title = "Sources"
	s.ui.RenderHTML(w, r, "sources.html", data)

}

// Handle adding new post via form
func (s *Service) NewSourceHandler(w http.ResponseWriter, r *http.Request) {

	// Compose data object
	data := ctxv.Get[*types.TemplateData](r.Context())

	// Populate needed data for an empty form
	data.Form = &types.Form{
		Legend: "New Playlist",
		Content: &types.FormGroup{
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

		var formError types.FlashMessage

		err := r.ParseForm()
		if err != nil {
			slog.WarnContext(
				r.Context(), "failed to parse the form",
				"error", err,
			)
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
			slog.WarnContext(
				r.Context(), "failed extract playlist ID",
				"error", err,
			)
			formError.Message = "Could not extract the playlist ID"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check if the playlist is already posted
		if s.sourcesRepo.SourceExists(r.Context(), playlistID) {
			slog.WarnContext(r.Context(), "playlist already posted")
			formError.Message = "Playlist already posted"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Fetch playlist metadata from YouTube
		sources, err := s.yt.GetSources(
			r.Context(),
			&retry.Config{
				MaxRetries: 3,
				MaxJitter:  time.Second,
				Delay:      time.Second,
			},
			playlistID,
		)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to get source metadata from YouTube",
				"error", err,
			)
			formError.Message = "Unable to fetch the playlist from YouTube"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Fetch channel data from YouTube
		channelID := sources[0].Snippet.ChannelId
		channels, err := s.yt.GetChannels(
			r.Context(),
			&retry.Config{
				MaxRetries: 3,
				MaxJitter:  time.Second,
				Delay:      time.Second,
			},
			channelID,
		)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to get channel metadata from YouTube",
				"error", err,
			)
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
			slog.ErrorContext(
				r.Context(), "failed to insert source in DB",
				"error", err,
			)
			formError.Message = "Could not create source"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the source
		redirectURL := fmt.Sprintf("/source/%s/", playlistID)
		redirectTo := redirect.Sanitize(redirectURL, auth.IsProtectedRoute)
		redirect.Execute(w, r, redirectTo, http.StatusFound)

	default:
		s.ui.HTMLError(w, r, data, http.StatusMethodNotAllowed)
	}
}

// Handle posts in a certain source
func (s *Service) SourcePostsHandler(w http.ResponseWriter, r *http.Request) {

	sourceID := r.PathValue("source")
	orderBy := r.URL.Query().Get("order_by")

	// Construct the Redis key
	redisKey := fmt.Sprintf("source:%s:posts", sourceID)

	switch orderBy {
	case types.Likes:
		redisKey += fmt.Sprintf(":%s", types.Likes)
	case types.AvgRating:
		redisKey += fmt.Sprintf(":%s", types.AvgRating)
	case types.RatingCount:
		redisKey += fmt.Sprintf(":%s", types.RatingCount)
	}

	// Generate template data
	data := ctxv.Get[*types.TemplateData](r.Context())

	var (
		err   error
		posts types.Posts
	)

	if data.CurrentUser.IsAdmin() {
		posts, err = s.postsRepo.GetSourcePosts(
			r.Context(), sourceID, "", orderBy,
		)
	} else {
		posts, err = rdb.GetCachedData(
			r.Context(),
			s.rdb,
			redisKey,
			s.config.CacheTimeout,
			func() (types.Posts, error) {
				return s.postsRepo.GetSourcePosts(
					r.Context(), sourceID, "", orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get source posts from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no source posts fetched from DB")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	data.Posts = &posts
	if sourceID == "other" {
		data.Posts.Title = "Other Uploads"
	}
	data.Title = data.Posts.Title
	s.ui.RenderHTML(w, r, "source.html", data)
}
