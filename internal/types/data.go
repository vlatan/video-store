package types

import (
	"html/template"

	"github.com/vlatan/video-store/internal/config"
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
	*config.Config
	*HTMLErrorData
	*PaginationInfo
	*Posts
	*Users
	*Form
}
