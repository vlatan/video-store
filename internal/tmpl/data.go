package tmpl

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/ctxv"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

type Base struct {
	*config.Config
	SearchQuery      string
	Categories       []types.Category
	CurrentURI       string
	BaseCanonicalURL string
	FlashMessages    []*types.FlashMessage
	CurrentUser      *types.User
}

type baseLoader struct {
	once sync.Once
	val  Base
}

// newBase creates new base data struct:
func (s *Service) newBase(w http.ResponseWriter, r *http.Request) Base {

	// Get the categories from cache
	categories, err := s.rdb.Get(
		r.Context(),
		"categories",
		s.cfg.CacheTimeout,
		func() (types.Categories, error) {
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

// Base uses newBase just once if loader in context, returns the base and
// stores the base in the loader for any subsequent calls.
func (s *Service) Base(w http.ResponseWriter, r *http.Request) Base {
	if l := ctxv.Get[*baseLoader](r.Context()); l != nil {
		l.once.Do(func() { l.val = s.newBase(w, r) })
		return l.val
	}
	return s.newBase(w, r)
}
