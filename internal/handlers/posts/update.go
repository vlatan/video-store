package posts

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/vlatan/video-store/internal/handlers/auth"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/normalize"
	"github.com/vlatan/video-store/internal/utils/redirect"
	"github.com/vlatan/video-store/internal/utils/stringx"
)

// UpdatePostHandler handles the post update
func (s *Service) UpdatePostHandler(w http.ResponseWriter, r *http.Request) {

	// Get video id from URL path
	videoID := r.PathValue("video")

	// Get template data
	data := s.ui.TmplData(w, r)

	// Validate the YT ID
	if validVideoID.FindStringSubmatch(videoID) == nil {
		slog.WarnContext(r.Context(), "invalid video id")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Get the post data straight from DB
	post, err := s.postsRepo.GetSinglePost(r.Context(), videoID)
	if errors.Is(err, pgx.ErrNoRows) {
		slog.WarnContext(
			r.Context(), "no such post in DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get the post from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	// Assign post data
	data.CurrentPost = &post
	if data.CurrentPost.Category == nil {
		data.CurrentPost.Category = &types.Category{}
	}

	// We need these for the release year in the form
	var released string
	maxYear := time.Now().Year() + 1
	decade := (maxYear % 100) / 10
	unit := maxYear % 10
	if data.CurrentPost.ReleaseYear != 0 {
		released = strconv.Itoa(int(data.CurrentPost.ReleaseYear))
	}

	// Populate needed data for the post form
	data.Form = &types.Form{
		Legend: "Edit Post",
		Title: &types.FormGroup{
			Label:       "Title",
			Placeholder: "Your title...",
			Value:       data.CurrentPost.GetTitle(),
		},
		Content: &types.FormGroup{
			Type:        types.FieldTypeTextarea,
			Label:       "Content",
			Placeholder: "You can use markdown...",
			Value:       data.CurrentPost.Summary,
		},
		Category: &types.FormGroup{
			Label: "Category",
			Value: data.CurrentPost.Category.Name,
		},
		ReleaseYear: &types.FormGroup{
			Label:       "Released",
			Placeholder: "YYYY",
			Value:       released,
			Pattern:     fmt.Sprintf(`(19\d\d|20[0-%d]\d|20%d[0-%d])`, decade-1, decade, unit),
			Title:       fmt.Sprintf("Year must be between 1900 and %d", maxYear),
		},
	}

	for _, director := range data.CurrentPost.Directors {
		data.Form.Directors = append(data.Form.Directors,
			&types.FormGroup{
				Label:       "Director",
				Placeholder: "Director's name...",
				Value:       director,
			},
		)
	}

	// Add one empty director input if none
	if len(data.Form.Directors) == 0 {
		data.Form.Directors = append(data.Form.Directors,
			&types.FormGroup{
				Label:       "Director",
				Placeholder: "Director's name...",
			},
		)
	}

	data.Title = "Edit This Post"

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

		// Get and assing back the values from the parsed form
		data.Form.Title.Value = r.FormValue("title")
		data.Form.Category.Value = r.FormValue("category")
		data.Form.Content.Value = r.FormValue("content")
		data.Form.ReleaseYear.Value = r.FormValue("released")

		directors := r.Form["directors"]
		data.Form.Directors = nil
		for _, director := range directors {
			data.Form.Directors = append(data.Form.Directors,
				&types.FormGroup{
					Label:       "Director",
					Placeholder: "Director's name...",
					Value:       director,
				},
			)
		}

		// Add one empty director input if none
		if len(data.Form.Directors) == 0 {
			data.Form.Directors = append(data.Form.Directors,
				&types.FormGroup{
					Label:       "Director",
					Placeholder: "Director's name...",
				},
			)
		}

		// Convert the release year to valid int16 year
		releaseYear, err := stringx.ParseReleaseYear(data.Form.ReleaseYear.Value)
		if err != nil {
			slog.WarnContext(
				r.Context(), "failed to parse release year",
				"error", err,
			)
			formError.Message = "Could not parse the release year"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Normalize the directors before DB upsert
		directors, err = normalize.Directors(directors)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to normalize directors",
				"error", err,
			)
			formError.Message = "Could not parse the directors"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Asign the new values to the current post
		data.CurrentPost.OriginalTitle = normalize.Title(
			data.Form.Title.Value,
			normalize.VideoTitleCutoffs,
		)
		data.CurrentPost.Category.Name = data.Form.Category.Value
		data.CurrentPost.Summary = normalize.Description(data.Form.Content.Value)
		data.CurrentPost.Directors = directors
		data.CurrentPost.ReleaseYear = releaseYear

		// Update the post
		rowsAffected, err := s.postsRepo.UpdatePost(r.Context(), data.CurrentPost)

		if err != nil || rowsAffected == 0 {
			slog.ErrorContext(
				r.Context(), "failed to update the post in DB",
				"error", err,
			)
			formError.Message = "Could not update the post in DB"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Delete the redis cache
		redisKey := fmt.Sprintf(postCacheKey, videoID)
		if err = s.rdb.Client.Del(r.Context(), redisKey).Err(); err != nil {
			slog.ErrorContext(
				r.Context(), "failed to delete the cache on post",
				"error", err,
			)
			formError.Message = "Could not delete the cache on post"
			data.Form.Error = &formError
			s.ui.RenderHTML(w, r, "form.html", data)
			return
		}

		// Check out the updated page
		redirectURL := fmt.Sprintf("/video/%s/", videoID)
		redirectTo := redirect.Sanitize(redirectURL, auth.IsProtectedRoute)
		redirect.Execute(w, r, redirectTo, http.StatusFound)

	default:
		s.ui.HTMLError(w, r, data, http.StatusMethodNotAllowed)
	}
}
