package ui

import (
	"log"
	"net/http"

	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"github.com/gorilla/csrf"
)

// NewData creates new default data struct to be passed to the templates
// Instead of manualy envoking this function in each route it can be envoked in a middleware
// and passed donwstream as value to the request context.
func (s *service) NewData(w http.ResponseWriter, r *http.Request) *models.TemplateData {

	// Get the categories from cache
	categories, _ := rdb.GetCachedData(
		r.Context(),
		s.rdb,
		"categories",
		s.config.CacheTimeout,
		func() (models.Categories, error) {
			return s.catsRepo.GetCategories(r.Context())
		},
	)

	// Construct the data
	data := &models.TemplateData{
		StaticFiles:  s.StaticFiles(),
		Config:       s.config,
		Categories:   categories,
		CurrentURI:   r.RequestURI,
		CanonicalURI: utils.CanonicalURI(r, s.config.Protocol),
		CSRFField:    csrf.TemplateField(r),
	}

	// Check if the path needs flash messages
	if utils.IsFilePath(r.URL.Path) {
		return data
	}

	// Check for flash cookie
	if _, err := r.Cookie(s.config.FlashSessionName); err != nil {
		return data
	}

	// Get any flash messages from session
	session, _ := s.store.Get(r, s.config.FlashSessionName)
	flashes := session.Flashes()

	var flashMessages []*models.FlashMessage
	for _, v := range flashes {
		if flash, ok := v.(*models.FlashMessage); ok && flash != nil {
			flashMessages = append(flashMessages, flash)
		}
	}

	// Clear the flash session created with s.store.Get
	session.Options.MaxAge = -1
	if err := session.Save(r, w); err != nil {
		log.Printf("unable to clear/save the flash session; %v", err)
	}

	// Put flash messages to data
	data.FlashMessages = flashMessages
	return data
}
