package ui

import (
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

// TmplData creates new default data struct to be passed to the templates
func (s *service) TmplData(w http.ResponseWriter, r *http.Request) *types.TemplateData {

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
	_, baseURL := pathx.CanonicalURLs(r, s.config.Protocol)

	// Construct the data
	data := &types.TemplateData{
		StaticFiles:      s.StaticFiles(),
		Config:           s.config,
		Categories:       categories,
		CurrentURI:       r.RequestURI,
		BaseCanonicalURL: baseURL,
		CurrentUser:      ctxd.GetUser(r.Context()),
	}

	// Check if the path needs flash messages
	if pathx.IsFile(r.URL.Path) || pathx.IsStatic(r.URL.Path) {
		return data
	}

	// Get flash messages from session and attach to data
	data.FlashMessages = s.session.Flashes(w, r)
	return data
}
