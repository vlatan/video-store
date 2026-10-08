package types

import (
	"html/template"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/config"
)

type TextFiles map[string]*FileInfo
type StaticFiles map[string]*FileInfo
type TemplateMap map[string]*template.Template

type FileInfo struct {
	Bytes      []byte
	Compressed []byte
	MediaType  string
	ModTime    time.Time
	Etag       string
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

// Data struct to pass to templates
type TemplateData struct {
	Title            string
	CurrentPost      *Post
	CurrentPage      *Page
	CurrentCategory  *Category
	CurrentSource    *Source
	CurrentYear      int16
	CurrentUser      *User
	CurrentURI       string
	BaseCanonicalURL string
	Sources          []Source
	Categories       []Category
	FlashMessages    []*FlashMessage
	SearchQuery      string
	XMLDeclarations  []template.HTML
	SitemapItems     []*SitemapItem
	StaticFiles
	*config.Config
	*HTMLErrorData
	*PaginationInfo
	*Posts
	*Users
	*Form
}

// Add version query string to file
func (td *TemplateData) AddVersion(path string) string {
	if fi, ok := td.StaticFiles[path]; ok && fi.Etag != "" {
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
