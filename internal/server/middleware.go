package server

import (
	"log"
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

		// Get user row ID from session
		id, ok := session.Values["ID"].(int)
		if !ok {
			next(w, r)
			return
		}

		// Update last seen
		now := time.Now()
		session.Values["LastSeen"] = now

		// Get lastSeenDB from session
		// This will be a zero time value (January 1, year 1, 00:00:00 UTC) on fail
		lastSeenDB := session.Values["LastSeenDB"].(time.Time)

		// Check if the DB update is out of sync for an entire date
		if !sameDate(lastSeenDB, now) {
			if err := s.db.UpdateUserLastSeen(id, now); err != nil {
				log.Printf("Couldn't update the last seen in DB on user with id '%d': %v\n", id, err)
			}
			session.Values["LastSeenDB"] = now
		}

		session.Save(r, w)
		next(w, r)
	}
}

func sameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}
