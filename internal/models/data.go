package models

import (
	"factual-docs/internal/handlers/files"
	"factual-docs/internal/shared/config"
	"html/template"
	"strings"
	"time"
)

// Flash message object to store to session for the next page
type FlashMessage struct {
	Message  string
	Category string
}

// Specific data for the error pages
type HTMLErrorData struct {
	Title   string
	Heading string
	Text    string
}

// Data struct to pass to templates
type TemplateData struct {
	StaticFiles   files.StaticFiles
	Config        *config.Config
	Title         string
	CurrentPost   *Post
	CurrentUser   *User
	CurrentURI    string
	CanonicalURL  string
	Posts         Posts
	Categories    []Category
	FlashMessages []*FlashMessage
	SearchQuery   string
	HTMLErrorData *HTMLErrorData
	CSRFField     template.HTML
}

// Check if current user is admin
func (td *TemplateData) IsCurrentUserAdmin() bool {
	return td.CurrentUser.IsAuthenticated() &&
		td.CurrentUser.UserID == td.Config.AdminOpenID
}

// Add version query string to file
func (td *TemplateData) AddVersion(path string) string {
	if fi, ok := td.StaticFiles[path]; ok {
		return path + "?v=" + fi.Etag
	}
	return path
}

// Split string helper function for templates
func (td *TemplateData) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Get time now
func (td *TemplateData) Now() time.Time {
	return time.Now()
}
