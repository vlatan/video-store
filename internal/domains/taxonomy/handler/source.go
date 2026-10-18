package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/domains/post"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/redirect"
	"github.com/vlatan/video-store/internal/utils/retry"
)

// Handle all sources page
func (s *Service) Sources(w http.ResponseWriter, r *http.Request) {

	var (
		err     error
		sources taxonomy.Sources
	)

	if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
		sources, err = s.taxonomy.Sources(r.Context())
	} else {
		sources, err = s.rdb.Get(
			r.Context(),
			"sources",
			s.cfg.CacheTimeout,
			func() (taxonomy.Sources, error) {
				return s.taxonomy.Sources(r.Context())
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get sources from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	if len(sources) == 0 {
		slog.WarnContext(r.Context(), "no sources found in DB")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Get template data
	data := tmplData{
		Title:   "Sources",
		Sources: sources,
		Base:    s.tmpl.Base(w, r),
	}

	s.tmpl.RenderHTML(w, r, "sources.html", data)

}

// Handle adding new post via form
func (s *Service) NewSource(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := tmplData{
		Title:  "Add New Source",
		Legend: "New Playlist",
		Content: &types.FormGroup{
			Label:       "Post YouTube Playlist URL",
			Placeholder: "Playlist URL here...",
		},
		Base: s.tmpl.Base(w, r),
	}

	switch r.Method {
	case "GET":
		// Serve the page with the form
		s.tmpl.RenderHTML(w, r, "source_form.html", data)

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
			s.tmpl.RenderHTML(w, r, "form.html", data)
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
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check if the playlist is already posted
		if s.taxonomy.SourceExists(r.Context(), playlistID) {
			slog.WarnContext(r.Context(), "playlist already posted")
			formError.Message = "Playlist already posted"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
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
			s.tmpl.RenderHTML(w, r, "form.html", data)
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
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Create a source object
		source := s.yt.NewYouTubeSource(sources[0], channels[0])
		source.UserID = data.CurrentUser.ID

		// Insert the source in DB
		rowsAffected, err := s.taxonomy.InsertSource(r.Context(), source)
		if err != nil || rowsAffected == 0 {
			slog.ErrorContext(
				r.Context(), "failed to insert source in DB",
				"error", err,
			)
			formError.Message = "Could not create source"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the source
		redirectTo := fmt.Sprintf("/source/%s/", playlistID)
		redirect.Safe(w, r, redirectTo, http.StatusFound)

	default:
		s.tmpl.HTMLError(w, r, http.StatusMethodNotAllowed)
	}
}

// Handle posts in a certain source
func (s *Service) SourcePosts(w http.ResponseWriter, r *http.Request) {

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

	var (
		err   error
		posts post.Posts
	)

	if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
		posts, err = s.post.Source(
			r.Context(), sourceID, "", orderBy,
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (post.Posts, error) {
				return s.post.Source(
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
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no source posts fetched from DB")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	if sourceID == "other" {
		posts.Title = "Other Uploads"
	}

	// Get template data
	data := tmplData{
		Title:         posts.Title,
		Posts:         posts,
		CurrentSource: &taxonomy.Source{PlaylistID: sourceID},
		Base:          s.tmpl.Base(w, r),
	}

	s.tmpl.RenderHTML(w, r, "taxonomy.html", data)
}

// DeleteSourceHandler deletes a source
func (s *Service) DeleteSource(w http.ResponseWriter, r *http.Request) {

	// Get the source ID from path
	playlistID := r.PathValue("source")
	rowsAffected, err := s.taxonomy.DeleteSource(r.Context(), playlistID)

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to delete the source",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		slog.WarnContext(r.Context(), "no such source to delete")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	successDelete := types.FlashMessage{
		Message:  "The source has been deleted!",
		Category: "info",
	}

	s.session.AddFlash(w, r, &successDelete)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Handle posts in a certain source
func (s *Service) SourcePostsAPI(w http.ResponseWriter, r *http.Request) {

	// Get the cursor from a query param
	cursor := r.URL.Query().Get("cursor")

	// Get the order_by query param if any
	orderBy := r.URL.Query().Get("order_by")

	// Get the playlist id
	sourceID := r.PathValue("source")

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

	if cursor != "" {
		redisKey += fmt.Sprintf(":cursor:%s", cursor)
	}

	// Get current user
	currentUser := ctxd.GetUser(r.Context())

	var (
		err   error
		posts post.Posts
	)

	if currentUser.IsAdmin(s.cfg) {
		posts, err = s.post.Source(
			r.Context(), sourceID, cursor, orderBy,
		)
	} else {
		posts, err = s.rdb.Get(
			r.Context(),
			redisKey,
			s.cfg.CacheTimeout,
			func() (post.Posts, error) {
				return s.post.Source(
					r.Context(), sourceID, cursor, orderBy,
				)
			},
		)
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(),
			"failed to get source posts from DB",
			"error", err,
		)
		s.tmpl.JSONError(w, r, http.StatusInternalServerError)
		return
	}

	if len(posts.Items) == 0 {
		slog.WarnContext(r.Context(), "no source posts found in DB")
		s.tmpl.JSONError(w, r, http.StatusNotFound)
		return
	}

	s.tmpl.WriteJSON(w, r, posts)
}
