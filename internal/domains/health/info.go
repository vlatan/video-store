package health

import "net/http"

// Info returns DB, Redis and server health status
func (s *Service) Info(w http.ResponseWriter, r *http.Request) {

	// Construct joined map
	data := map[string]any{
		"redis_status":    s.rdb.Health(r.Context()),
		"database_status": s.db.Health(r.Context()),
		"server_status":   getServerStats(),
	}

	w.Header().Set("X-Robots-Tag", "noindex")
	s.tmpl.WriteJSON(w, r, data)
}
