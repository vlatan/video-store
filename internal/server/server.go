package server

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"factual-docs/internal/auth"
	"factual-docs/internal/posts"
	"factual-docs/internal/services/config"
	"factual-docs/internal/services/database"
	"factual-docs/internal/services/files"
	"factual-docs/internal/services/redis"
	tmpls "factual-docs/internal/services/templates"
	"factual-docs/internal/users"

	"github.com/gorilla/sessions"
)

type Server struct {
	config *config.Config
	store  *sessions.CookieStore
	db     database.Service
	rdb    redis.Service
	tm     tmpls.Service
	sf     files.StaticFiles
	users  *users.Service
	auth   *auth.Service
	posts  *posts.Service
}

// Create new HTTP server
func NewServer() *http.Server {

	// Register types with gob to be able to use them in sessions
	gob.Register(&tmpls.FlashMessage{})
	gob.Register(time.Time{})

	cfg := config.New()     // Create new config service
	db := database.New(cfg) // Create database service
	rdb := redis.New(cfg)   // Create Redis service

	users := users.New(db)                   // Create users service
	store := NewCookieStore(cfg)             // Create cookie store
	auth := auth.New(users, store, rdb, cfg) // Create auth service

	sf := files.New()                        // Create minified files map
	tm := tmpls.New(db, rdb, cfg, store, sf) // Create parsed templates map

	// Create new Server struct
	newServer := &Server{
		config: cfg,
		sf:     sf,
		rdb:    rdb,
		db:     db,
		store:  store,
		tm:     tm,
		users:  users,
		auth:   auth,
		posts:  posts.New(db, rdb, tm, cfg, auth),
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
