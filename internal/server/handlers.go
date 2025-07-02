package server

import (
	"fmt"
	"log"
	"net/http"
)

// Handle ads.txt
func (s *Server) adsTextHandler(w http.ResponseWriter, r *http.Request) {
	if s.config.AdSenseAccount == "" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	content := fmt.Sprintf("google.com, pub-%s, DIRECT, f08c47fec0942fa0", s.config.AdSenseAccount)
	if _, err := w.Write([]byte(content)); err != nil {
		log.Printf("Failed to write response to '/ads.txt': %v", err)
	}
}

// DB and Redis health status
// Wrap this with middlware that allows only admins
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {

	// Construct joined map
	data := map[string]any{
		"redis_status":    s.rdb.Health(r.Context()),
		"database_status": s.db.Health(r.Context()),
		"server_status":   getServerStats(),
	}

	s.tm.WriteJSON(w, r, data)
}
