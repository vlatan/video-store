package server

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"factual-docs/internal/config"
	"factual-docs/internal/database"
	"factual-docs/internal/files"
	"factual-docs/internal/redis"
	"factual-docs/internal/templates"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

type Server struct {
	config *config.Config
	store  *sessions.CookieStore
	db     database.Service
	rdb    redis.Service
	tm     templates.Service
	sf     files.StaticFiles
}

// Create new HTTP server
func NewServer() *http.Server {

	// Register types with gob to be able to use them in sessions
	gob.Register(&templates.FlashMessage{})
	gob.Register(time.Time{})

	// Create new config object
	cfg := config.New()

	// Create new Server struct
	newServer := &Server{
		config: cfg,
		sf:     files.New(),         // Create minified files map
		rdb:    redis.New(cfg),      // Create Redis service
		db:     database.New(cfg),   // Create database service
		store:  NewCookieStore(cfg), // Create cookie store
		tm:     templates.New(),     // Create parsed templates map
	}

	// Declare Server config
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
}

// Creates new default data struct to be passed to the templates
// Instead of manualy envoking this function in each route it can be envoked in a middleware
// and passed donwstream as value to the request context.
func (s *Server) NewData(w http.ResponseWriter, r *http.Request) *templates.TemplateData {

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
	flashMessages := []*templates.FlashMessage{}
	for _, v := range flashes {
		if flash, ok := v.(*templates.FlashMessage); ok && flash != nil {
			flashMessages = append(flashMessages, flash)
		}
	}
	session.Save(r, w)

	return &templates.TemplateData{
		StaticFiles:   s.sf,
		Config:        s.config,
		Categories:    categories,
		CurrentURI:    r.RequestURI,
		FlashMessages: flashMessages,
		CSRFField:     csrf.TemplateField(r),
	}
}

// Get basic server stats
func getServerStats() map[string]any {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]any{
		"goroutines":   runtime.NumGoroutine(),
		"memory_alloc": m.Alloc,
		"memory_sys":   m.Sys,
		"gc_runs":      m.NumGC,
		"cpu_count":    runtime.NumCPU(),
	}
}
