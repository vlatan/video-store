package ui

import (
	"log"
	"math"
	"net/http"

	"github.com/vlatan/video-store/internal/drivers/redis"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"

	"github.com/gorilla/csrf"
)

// Get the map containing the static files
func (s *service) GetStaticFiles() models.StaticFiles {
	return s.staticFiles
}

// Creates new default data struct to be passed to the templates
// Instead of manualy envoking this function in each route it can be envoked in a middleware
// and passed donwstream as value to the request context.
func (s *service) NewData(w http.ResponseWriter, r *http.Request) *models.TemplateData {

	// Get the categories from cache
	categories, _ := redis.GetItems(
		true,
		r.Context(),
		s.rdb,
		"categories",
		s.config.CacheTimeout,
		func() ([]models.Category, error) {
			return s.catsRepo.GetCategories(r.Context())
		},
	)

	// Construct the data
	data := &models.TemplateData{
		StaticFiles: s.GetStaticFiles(),
		Config:      s.config,
		Categories:  categories,
		CurrentURI:  r.RequestURI,
		BaseURL:     utils.GetBaseURL(r, s.config.Protocol),
		CSRFField:   csrf.TemplateField(r),
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

// Create new pagination struct that contains all the info about the pagination element
func (s *service) NewPagination(currentPage, totalRecords, pageSize int) *models.PaginationInfo {

	// Total pages are the celing division between total records and the records on one page
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))
	// Avoid zero or negative  values
	totalPages = max(totalPages, 1)

	// Avoid zero or negative  values
	currentPage = max(currentPage, 1)
	// Current page can't be greater than total pages
	currentPage = min(currentPage, totalPages)

	return &models.PaginationInfo{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
		PageSize:     pageSize,
		Pages:        generatePageNumbers(currentPage, totalPages),
	}
}

// Creates the page number sequence with ellipsis
func generatePageNumbers(currentPage, totalPages int) (pages []models.PageInfo) {

	// No pages if just one page
	if totalPages <= 1 {
		return pages
	}

	// Always show the first page
	pages = append(pages, models.PageInfo{
		Number:     1,
		IsCurrent:  currentPage == 1,
		IsEllipsis: false,
	})

	// The range of pages to show around the current page
	start := max(2, currentPage-1)
	end := min(totalPages-1, currentPage+1)

	// Add ellipsis after first page if needed
	if start > 2 {
		pages = append(pages, models.PageInfo{
			IsCurrent:  false,
			IsEllipsis: true,
		})
	}

	// Add the range of pages around current page
	for i := start; i <= end; i++ {
		pages = append(pages, models.PageInfo{
			Number:     i,
			IsCurrent:  i == currentPage,
			IsEllipsis: false,
		})
	}

	// Add ellipsis before last page if needed
	if end < totalPages-1 {
		pages = append(pages, models.PageInfo{
			IsCurrent:  false,
			IsEllipsis: true,
		})
	}

	// Always show last page (if it's not page 1)
	if totalPages > 1 {
		pages = append(pages, models.PageInfo{
			Number:     totalPages,
			IsCurrent:  currentPage == totalPages,
			IsEllipsis: false,
		})
	}

	return pages
}
