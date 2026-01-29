package ui

import (
	"math"

	"github.com/vlatan/video-store/internal/models"
)

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
