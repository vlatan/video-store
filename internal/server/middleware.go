package server

import (
	"net/http"
	"time"
)

// Record the last seen date of the user
func (s *Server) userLastSeen(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session
		session, err := s.store.Get(r, s.config.SessionName)
		if session == nil || err != nil {
			next(w, r)
			return
		}

		// Get userID from session
		_, exists := session.Values["UserID"].(string)
		if !exists {
			next(w, r)
			return
		}

		// Get lastSeenDB from session
		lastSeenDB, exists := session.Values["LastSeen"].(time.Time)
		if !exists {
			next(w, r)
			return
		}

		// Update last seen
		now := time.Now()
		session.Values["LastSeen"] = now

		// Check if the DB update is out of sync of a whole day
		if lastSeenDB.IsZero() || !sameDate(lastSeenDB, now) {
			// err := updateUserLastSeenInDB(userID)
			session.Values["LastSeen"] = now
		}

		// if session.UserID > 0 { // user is logged in
		// 	updateLastSeen(w, r, session)
		// }

		next(w, r)
	}
}

func sameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// func (s *Server) updateUserLastSeen(userID int) error {
// 	// TODO
// }
