package models

import "math"

type PaginationInfo struct {
	CurrentPage  int
	TotalPages   int
	TotalRecords int
	PageSize     int
	HasPrevious  bool
	HasNext      bool
	Pages        []PageInfo
}

type PageInfo struct {
	Number     int
	IsCurrent  bool
	IsEllipsis bool
}

// Create new pagination struct that contains all the info about the pagination element
func CalculatePagination(currentPage, totalRecords, pageSize int) *PaginationInfo {

	// Total pages are the celing division between total records and the records on one page
	totalPages := int(math.Ceil(float64(totalRecords) / float64(pageSize)))
	// Avoid zero or negative  values
	totalPages = max(totalPages, 1)

	// Avoid zero or negative  values
	currentPage = max(currentPage, 1)
	// Current page can't be greater than total pages
	currentPage = min(currentPage, totalPages)

	return &PaginationInfo{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		TotalRecords: totalRecords,
		PageSize:     pageSize,
		Pages:        generatePageNumbers(currentPage, totalPages),
	}
}

// Creates the page number sequence with ellipsis
func generatePageNumbers(currentPage, totalPages int) (pages []PageInfo) {

	// If we have 7 or fewer pages, show them all
	if totalPages <= 7 {
		for i := 1; i <= totalPages; i++ {
			pages = append(pages, PageInfo{
				Number:     i,
				IsCurrent:  i == currentPage,
				IsEllipsis: false,
			})
		}
		return pages
	}

	// Always show the first page
	pages = append(pages, PageInfo{
		Number:     1,
		IsCurrent:  currentPage == 1,
		IsEllipsis: false,
	})

	// Determine the range of pages to show around current page
	start := max(2, currentPage-2)
	end := min(totalPages-1, currentPage+2)

	// Add ellipsis after first page if needed
	if start > 2 {
		pages = append(pages, PageInfo{
			Number:     0,
			IsCurrent:  false,
			IsEllipsis: true,
		})
	}

	// Add the range of pages around current page
	for i := start; i <= end; i++ {
		pages = append(pages, PageInfo{
			Number:     i,
			IsCurrent:  i == currentPage,
			IsEllipsis: false,
		})
	}

	// Add ellipsis before last page if needed
	if end < totalPages-1 {
		pages = append(pages, PageInfo{
			Number:     0,
			IsCurrent:  false,
			IsEllipsis: true,
		})
	}

	// Always show last page (if it's not page 1)
	if totalPages > 1 {
		pages = append(pages, PageInfo{
			Number:     totalPages,
			IsCurrent:  currentPage == totalPages,
			IsEllipsis: false,
		})
	}

	return pages
}
