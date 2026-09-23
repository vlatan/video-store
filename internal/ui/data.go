package ui

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/paths"
)

// NewData creates new default data struct to be passed to the templates
// Instead of manualy envoking this function in each route it can be envoked in a middleware
// and passed donwstream as value to the request context.
func (s *service) NewTemplateData(w http.ResponseWriter, r *http.Request) *types.TemplateData {

	// Get the categories from cache
	categories, _ := rdb.GetCachedData(
		r.Context(),
		s.rdb,
		"categories",
		s.config.CacheTimeout,
		func() (types.Categories, error) {
			return s.catsRepo.GetCategories(r.Context())
		},
	)

	// Get the base canonical URL without queries and fragments
	_, baseURL := paths.CanonicalURLs(r, s.config.Protocol)

	// Construct the data
	data := &types.TemplateData{
		StaticFiles:      s.StaticFiles(),
		Config:           s.config,
		Categories:       categories,
		CurrentURI:       r.RequestURI,
		BaseCanonicalURL: baseURL,
	}

	// Check if the path needs flash messages
	if paths.IsFile(r.URL.Path) {
		return data
	}

	// Check for flash cookie
	if _, err := r.Cookie(s.config.FlashSessionName); err != nil {
		return data
	}

	// Get any flash messages from session
	session, _ := s.store.Get(r, s.config.FlashSessionName)
	flashes := session.Flashes()

	var flashMessages []*types.FlashMessage
	for _, v := range flashes {
		if flash, ok := v.(*types.FlashMessage); ok && flash != nil {
			flashMessages = append(flashMessages, flash)
		}
	}

	// Clear the flash session created with s.store.Get
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to clear the flash session",
			"error", err,
		)
	}

	// Put flash messages to data
	data.FlashMessages = flashMessages
	return data
}
