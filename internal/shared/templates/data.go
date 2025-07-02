package tmpls

import (
	"factual-docs/internal/shared/config"
	"factual-docs/internal/shared/database"
	"factual-docs/internal/shared/files"
	"factual-docs/internal/shared/redis"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/csrf"
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

// Creates new default data struct to be passed to the templates
// Instead of manualy envoking this function in each route it can be envoked in a middleware
// and passed donwstream as value to the request context.
func (s *service) NewData(w http.ResponseWriter, r *http.Request) *TemplateData {

	var categories []database.Category
	redis.GetItems(
		true,
		r.Context(),
		s.rdb,
		"categories",
		s.config.CacheTimeout,
		&categories,
		func() ([]database.Category, error) {
			return s.db.GetCategories(r.Context())
		},
	)

	// Get any flash messages from session and put to data
	session, _ := s.store.Get(r, s.config.FlashSessionName)
	flashes := session.Flashes()
	flashMessages := []*FlashMessage{}
	for _, v := range flashes {
		if flash, ok := v.(*FlashMessage); ok && flash != nil {
			flashMessages = append(flashMessages, flash)
		}
	}
	session.Save(r, w)

	return &TemplateData{
		StaticFiles:   s.sf.GetStaticFiles(),
		Config:        s.config,
		Categories:    categories,
		CurrentURI:    r.RequestURI,
		CanonicalURL:  getCanonicalURL(r),
		FlashMessages: flashMessages,
		CSRFField:     csrf.TemplateField(r),
	}
}

// Get canonilca absolute URL
func getCanonicalURL(r *http.Request) string {
	// Determine scheme
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	canonical := &url.URL{
		Scheme: scheme,
		Host:   r.Host,
		Path:   r.URL.Path,
	}

	return canonical.String()
}
