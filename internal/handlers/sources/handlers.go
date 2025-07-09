package sources

import (
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"log"
	"net/http"
)

// Handle all sources page
func (s *Service) SourcesHandler(w http.ResponseWriter, r *http.Request) {
	// Generate template data
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

	// Get sources from redis or DB
	var sources []models.Source
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		"all:sources",
		s.config.CacheTimeout,
		&sources,
		func() ([]models.Source, error) {
			return s.sourcesRepo.GetSources(r.Context())
		},
	)

	if err != nil {
		log.Printf("Was unabale to fetch sources on URI '%s': %v", r.RequestURI, err)
		s.tm.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if len(sources) == 0 {
		log.Printf("Fetched zero sources on URI '%s'", r.RequestURI)
		s.tm.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	data.Sources = sources
	data.Title = "Sources"
	s.tm.RenderHTML(w, r, "sources", data)

}
