package tmpls

import (
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"net/http"
	"net/url"

	"github.com/gorilla/csrf"
)

// Creates new default data struct to be passed to the templates
// Instead of manualy envoking this function in each route it can be envoked in a middleware
// and passed donwstream as value to the request context.
func (s *service) NewData(w http.ResponseWriter, r *http.Request) *models.TemplateData {

	var categories []models.Category
	redis.GetItems(
		true,
		r.Context(),
		s.rdb,
		"categories",
		s.config.CacheTimeout,
		&categories,
		func() ([]models.Category, error) {
			return s.catRepo.GetCategories(r.Context())
		},
	)

	// Get any flash messages from session and put to data
	session, _ := s.store.Get(r, s.config.FlashSessionName)
	flashes := session.Flashes()
	flashMessages := []*models.FlashMessage{}
	for _, v := range flashes {
		if flash, ok := v.(*models.FlashMessage); ok && flash != nil {
			flashMessages = append(flashMessages, flash)
		}
	}
	session.Save(r, w)

	return &models.TemplateData{
		StaticFiles:   s.sf.GetStaticFiles(),
		Config:        s.config,
		Categories:    categories,
		CurrentURI:    r.RequestURI,
		CanonicalURL:  getCanonicalURL(r),
		FlashMessages: flashMessages,
		CSRFField:     csrf.TemplateField(r),
	}
}

// Get canonilca absolute URL
func getCanonicalURL(r *http.Request) string {
	// Determine scheme
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	canonical := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   r.URL.Path,
	}

	return canonical.String()
}
