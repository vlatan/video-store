package health

import (
	"net/http"

	"github.com/vlatan/video-store/internal/drivers/database"
	"github.com/vlatan/video-store/internal/drivers/rdb"
	"github.com/vlatan/video-store/internal/ui"
)

type Service struct {
	db  *database.Service
	rdb *rdb.Service
	ui  ui.Service
}

func New(db *database.Service, rdb *rdb.Service, ui ui.Service) *Service {
	return &Service{
		db:  db,
		rdb: rdb,
		ui:  ui,
	}
}

// DB and Redis health status
// Wrap this with middlware that allows only admins
func (s *Service) API(w http.ResponseWriter, r *http.Request) {

	// Construct joined map
	data := map[string]any{
		"redis_status":    s.rdb.Health(r.Context()),
		"database_status": s.db.Health(r.Context()),
		"server_status":   getServerStats(),
	}

	s.ui.WriteJSON(w, r, data)
}
