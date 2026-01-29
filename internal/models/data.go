package models

import (
	"html/template"
	"net/url"
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

// The response from the Genai API
type GenaiResponse struct {
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Category string `json:"category"`
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

type FieldType int

const (
	FieldTypeInput FieldType = iota
	FieldTypeTextarea
)

type FormGroup struct {
	Type        FieldType
	Label       string
	Placeholder string
	Value       string
}

// Returns true if the field type is input
func (ft FieldType) IsInput() bool {
	return ft == FieldTypeInput
}

// Returns true if the field type is textarea
func (ft FieldType) IsTextarea() bool {
	return ft == FieldTypeTextarea
}

type Form struct {
	Legend  string
	Title   *FormGroup
	Content *FormGroup
	Error   *FlashMessage
}

// Data struct to pass to templates
type TemplateData struct {
	StaticFiles     StaticFiles
	Config          *config.Config
	Title           string
	CurrentPost     *Post
	CurrentPage     *Page
	CurrentUser     *User
	CurrentURI      string
	BaseURL         *url.URL
	Sources         []Source
	Categories      []Category
	FlashMessages   []*FlashMessage
	SearchQuery     string
	CSRFField       template.HTML
	XMLDeclarations []template.HTML
	SitemapItems    []*SitemapItem
	*HTMLErrorData
	*PaginationInfo
	*Posts
	*Users
	*Form
}

func (td *TemplateData) CanonicalURL() string {
	return td.BaseURL.String()
}

func (td *TemplateData) AbsoluteURL(path string) string {
	u := *td.BaseURL // Copy the URL
	u.Path = path
	return u.String()
}

// Check if current user is admin
func (td *TemplateData) IsCurrentUserAdmin() bool {
	return td != nil && td.CurrentUser.IsAdmin(
		td.Config.AdminProviderUserId,
		td.Config.AdminProvider,
	)
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
