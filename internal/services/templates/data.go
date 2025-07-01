package templates

import (
	"factual-docs/internal/services/config"
	"factual-docs/internal/services/database"
	"factual-docs/internal/services/files"
	"html/template"
	"strings"
	"time"
)

// User struct to store in the USER info in session
// A simplified version of goth.User
type User struct {
	ID             int    `json:"id"`
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Provider       string `json:"provider"`
	AvatarURL      string `json:"avatar_url"`
	AnalyticsID    string `json:"analytics_id"`
	LocalAvatarURL string `json:"local_avatar_url"`
	AccessToken    string `json:"access_token"`
}

func (u *User) IsAuthenticated() bool {
	return u != nil && u.UserID != ""
}

type FlashMessage struct {
	Message  string
	Category string
}

type HTMLErrorData struct {
	Title   string
	Heading string
	Text    string
}

type TemplateData struct {
	StaticFiles   files.StaticFiles
	Config        *config.Config
	Title         string
	CurrentPost   *database.Post
	CurrentUser   *User
	CurrentURI    string
	CanonicalURL  string
	Posts         database.Posts
	Categories    []database.Category
	FlashMessages []*FlashMessage
	SearchQuery   string
	HTMLErrorData *HTMLErrorData
	CSRFField     template.HTML
}

func (td *TemplateData) IsCurrentUserAdmin() bool {
	return td.CurrentUser.IsAuthenticated() &&
		td.CurrentUser.UserID == td.Config.AdminOpenID
}

func (td *TemplateData) AddVersion(path string) string {
	if fi, ok := td.StaticFiles[path]; ok {
		return path + "?v=" + fi.Etag
	}
	return path
}

func (td *TemplateData) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func (td *TemplateData) Now() time.Time {
	return time.Now()
}
