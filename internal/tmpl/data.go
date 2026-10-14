package tmpl

import (
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

// Data creates new default data struct to be passed to the templates
func (s *Service) Data(w http.ResponseWriter, r *http.Request) *types.TemplateData {

	// Get the categories from cache
	categories, _ := s.rdb.Get(
		r.Context(),
		"categories",
		s.cfg.CacheTimeout,
		func() (types.Categories, error) {
			return s.taxonomy.Categories(r.Context())
		},
	)

	// Get the base canonical URL without queries and fragments
	_, baseURL := pathx.CanonicalURLs(r, s.cfg.Protocol)

	// Construct the data
	data := &types.TemplateData{
		Config:           s.cfg,
		Categories:       categories,
		CurrentURI:       r.RequestURI,
		BaseCanonicalURL: baseURL,
		CurrentUser:      ctxd.GetUser(r.Context()),
	}

	return data
}
