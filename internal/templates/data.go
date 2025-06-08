package templates

import (
	"factual-docs/internal/config"
	"factual-docs/internal/database"
	"factual-docs/internal/files"
	"net/url"
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
}

func (u *User) IsAuthenticated() bool {
	return u != nil && u.UserID != ""
}

type FlashMessage struct {
	Message  string
	Category string
}

type TemplateData struct {
	StaticFiles   files.StaticFiles
	Config        *config.Config
	Title         string
	SinglePost    database.Post
	Posts         []database.Post
	Categories    []database.Category
	CurrentUser   *User
	CurrentURI    string
	FlashMessages []*FlashMessage
}

func (td *TemplateData) IsAdmin() bool {
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

func (td *TemplateData) CurrentURLPath() string {
	u, err := url.Parse(td.CurrentURI)
	if err != nil {
		return td.CurrentURI
	}
	return u.Path
}
