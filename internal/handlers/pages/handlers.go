package pages

import (
	"errors"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/redis"
	"factual-docs/internal/shared/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// Handle single page
func (s *Service) SinglePageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug from URL
	slug := r.PathValue("slug")

	// Generate the default data
	data := s.ui.NewData(w, r)
	data.CurrentUser = utils.GetUserFromContext(r)

	var page models.Page
	err := redis.GetItems(
		!data.IsCurrentUserAdmin(),
		r.Context(),
		s.rdb,
		fmt.Sprintf("page:%s", slug),
		s.config.CacheTimeout,
		&page,
		func() (models.Page, error) {
			return s.pagesRepo.GetSinglePage(r.Context(), slug)
		},
	)

	if errors.Is(err, pgx.ErrNoRows) {
		log.Println("Can't find the page in DB:", slug)
		s.ui.HTMLError(w, r, http.StatusNotFound, data)
		return
	}

	if err != nil {
		log.Printf("Error while getting the page '%s' from DB: %v", slug, err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	// Assign the page to data
	data.CurrentPage = &page
	data.Title = page.Title

	s.ui.RenderHTML(w, r, "page.html", data)
}
