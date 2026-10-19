package tmpl

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/domains/taxonomy"
	"github.com/vlatan/video-store/internal/domains/user"
	"github.com/vlatan/video-store/internal/drivers/store"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

type Base struct {
	*config.Config
	SearchQuery      string
	Categories       []taxonomy.Category
	CurrentURI       string
	BaseCanonicalURL string
	FlashMessages    []*store.FlashMessage
	CurrentUser      *user.User
}

// newBase creates new base data struct:
func (s *Service) Base(w http.ResponseWriter, r *http.Request) Base {

	// Get the categories from cache
	categories, err := s.rdb.Get(
		r.Context(),
		"categories",
		s.cfg.CacheTimeout,
		func() (taxonomy.Categories, error) {
			return s.taxonomy.Categories(r.Context())
		},
	)

	if err != nil {
		slog.WarnContext(
			r.Context(),
			"failed to get categories", "error", err,
		)
	}

	// Get the base canonical URL without queries and fragments
	_, baseURL := pathx.CanonicalURLs(r, s.cfg.Protocol)

	// Construct the data
	return Base{
		Config:           s.cfg,
		Categories:       categories,
		CurrentURI:       r.RequestURI,
		BaseCanonicalURL: baseURL,
		FlashMessages:    s.session.Flashes(w, r),
		CurrentUser:      ctxd.GetUser(r.Context()),
	}
}
