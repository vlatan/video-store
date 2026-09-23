package utils

import (
	"log"
	"net/http"
	"strconv"
)

// Get page number from the request query param
// Defaults to 1 if invalid page
func GetPageNum(r *http.Request) (page int) {
	pageStr := r.URL.Query().Get("page")
	if pageInt, err := strconv.Atoi(pageStr); err == nil {
		page = pageInt
	}

	return max(page, 1)
}

// LogPlainln prints a line without a prefix using the log package
func LogPlainln(v ...any) {
	flags := log.Flags()
	log.SetFlags(0)
	log.Println(v...)
	log.SetFlags(flags)
}
