package posts

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/redirect"
	"github.com/vlatan/video-store/internal/utils/retry"
	"github.com/vlatan/video-store/internal/utils/sleep"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

// Handle adding new post via form
func (s *Service) New(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := postData{
		Title:  "Add New Video",
		Legend: "New Video",
		Content: &types.FormGroup{
			Label:       "Post YouTube Video URL",
			Placeholder: "Video URL here...",
		},
		Base: s.tmpl.Base(w, r),
	}

	switch r.Method {
	case "GET":
		// Serve the page with the form
		s.tmpl.RenderHTML(w, r, "post_form.html", data)

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
		videoID, err := extractYouTubeID(url)
		if err != nil {
			slog.WarnContext(
				r.Context(), "failed to extract the video ID",
				"error", err,
			)
			formError.Message = "Could not extract the video ID"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Validate the YT ID
		if validVideoID.FindStringSubmatch(videoID) == nil {
			slog.WarnContext(r.Context(), "invalid video id")
			formError.Message = "Could not validate the video ID"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check if the video is already posted
		if err = s.post.Exists(r.Context(), videoID); err == nil {
			slog.WarnContext(r.Context(), "video already posted")
			formError.Message = "Video already posted"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Fetch video data from YouTube
		metadata, err := s.yt.GetVideos(
			r.Context(),
			&retry.Config{
				MaxRetries: 3,
				MaxJitter:  time.Second,
				Delay:      time.Second,
			},
			videoID,
		)

		if err != nil {
			slog.WarnContext(
				r.Context(), "failed get video data from YouTube",
				"error", err,
			)
			formError.Message = "Unable to fetch the video from YouTube"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Validate the video data
		if err := s.yt.ValidateYouTubeVideo(metadata[0]); err != nil {
			slog.WarnContext(
				r.Context(), "failed to validate the video data",
				"error", err,
			)
			formError.Message = stringx.Capitalize(err.Error())
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Create post object
		post := s.yt.NewYouTubePost(metadata[0], "")
		post.UserActions = types.Actions{UserID: data.CurrentUser.ID}

		// Insert the video
		rowsAffected, err := s.post.Insert(r.Context(), post)
		if err != nil || rowsAffected == 0 {
			slog.WarnContext(
				r.Context(), "failed to insert the post in DB",
				"error", err,
			)
			formError.Message = "Could not insert the video in DB"
			data.Form.Error = &formError
			s.tmpl.RenderHTML(w, r, "form.html", data)
			return
		}

		// Generate content in the background using Gemini.
		// In production no need to use it, the worker will
		// generate the post content overnight.
		go func() {

			// Detach the request context and
			// give this goroutine reasonable time to finish
			detachedCtx := context.WithoutCancel(r.Context())
			ctx, cancel := context.WithTimeout(detachedCtx, 30*time.Minute)
			defer cancel()

			// Just give it a one try
			retryConfig := &retry.Config{
				MaxRetries: 1,
				MaxJitter:  2 * time.Second,
				Delay:      65 * time.Second,
			}

			if err := s.gemini.GeneratePostSummary(ctx, post, retryConfig); err != nil {
				slog.WarnContext(
					r.Context(),
					"failed to generate/update LLM post summary/category",
					"error", err,
				)
				// Exit early, do not try OCR nor DB update.
				return
			}

			// Sleep with context in mind for 60-90 seconds.
			// Min sleep needs to be 60s to avoid the genai 250k TPM quota.
			if err := sleep.Jitter(ctx, 60*time.Second, 90*time.Second); err != nil {
				return
			}

			if err := s.gemini.GeneratePostOCR(ctx, post, retryConfig); err != nil {
				slog.WarnContext(
					r.Context(),
					"failed to generate/update LLM post OCR data",
					"error", err,
				)
			}

			_, err = s.post.Update(ctx, post)
			if err != nil {
				slog.WarnContext(
					r.Context(),
					"failed to update LLM content in DB",
					"error", err,
				)
			}
		}()

		// Check out the video
		redirectTo := fmt.Sprintf("/video/%s/", videoID)
		redirect.Safe(w, r, redirectTo, http.StatusFound)

	default:
		s.tmpl.HTMLError(w, r, http.StatusMethodNotAllowed)
	}
}
