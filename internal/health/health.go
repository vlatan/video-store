package health

import (
	"net/http"

	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/tmpl"
)

type Service struct {
	db   *database.Service
	rdb  *rdb.Service
	tmpl *tmpl.Service
}

func New(db *database.Service, rdb *rdb.Service, tmpl *tmpl.Service) *Service {
	return &Service{
		db:   db,
		rdb:  rdb,
		tmpl: tmpl,
	}
}

// Handle handles DB and Redis health status
func (s *Service) Handle(w http.ResponseWriter, r *http.Request) {

	// Construct joined map
	data := map[string]any{
		"redis_status":    s.rdb.Health(r.Context()),
		"database_status": s.db.Health(r.Context()),
		"server_status":   getServerStats(),
	}

	s.tmpl.WriteJSON(w, r, data)
}
