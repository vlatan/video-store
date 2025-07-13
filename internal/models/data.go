package models

import (
	"factual-docs/internal/handlers/static"
	"factual-docs/internal/shared/config"
	"html/template"
	"strings"
	"time"
)

// The response from the Genai API
type GenaiResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

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

// Specific data for the JSON response
type JSONErrorData struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

type FormGroup struct {
	Label       string
	Placeholder string
	Value       string
}

func (f *FormGroup) IsEmpty() bool {
	return f.Label == "" && f.Placeholder == ""
}

type Form struct {
	Legend  string
	Content FormGroup
	Error   *FlashMessage
}

// Data struct to pass to templates
type TemplateData struct {
	StaticFiles  static.StaticFiles
	Config       *config.Config
	Title        string
	CurrentPost  *Post
	CurrentUser  *User
	CurrentURI   string
	CanonicalURL string
	*Posts
	Sources         []Source
	Categories      []Category
	FlashMessages   []*FlashMessage
	SearchQuery     string
	HTMLErrorData   *HTMLErrorData
	CSRFField       template.HTML
	XMLDeclarations []template.HTML
	*Form
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
