package templates

import (
	"factual-docs/internal/config"
	"factual-docs/internal/database"
	"factual-docs/internal/files"
	"strings"
	"time"
)

// User struct to store in the USER info in session
// A simplified version of goth.User
type User struct {
	UserID    string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Provider  string `json:"provider"`
	AvatarURL string `json:"avatar_url"`
}

type TemplateData struct {
	StaticFiles files.StaticFiles
	Config      *config.Config
	Title       string
	Posts       *[]database.Post
	Categories  *[]database.Category
	CurrentUser *User
}

func (td *TemplateData) StaticUrl(path string) string {
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
