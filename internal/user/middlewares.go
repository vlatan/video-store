package user

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

// IsAuth checks if the user is authenticated
func (s *Service) IsAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// If the user is authenticated move onto the next handler
		if ctxd.GetUser(r.Context()).IsAuthenticated() {
			next(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			s.tmpl.JSONError(w, r, http.StatusForbidden)
			return
		}

		data := s.tmpl.Data(w, r)
		s.tmpl.HTMLError(w, r, data, http.StatusForbidden)
	}
}

// IsAdmin checks if the user is admin
func (s *Service) IsAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// If the user is admin move onto the next handler
		if ctxd.GetUser(r.Context()).IsAdmin(s.cfg) {
			next(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/api/") {
			s.tmpl.JSONError(w, r, http.StatusForbidden)
			return
		}

		data := s.tmpl.Data(w, r)
		s.tmpl.HTMLError(w, r, data, http.StatusForbidden)
	}
}

// Load gets the user id and data from session and DB and stores it in the context
func (s *Service) Load(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip static files, they don't need user
		if pathx.IsStatic(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Fetch the user from session
		user := s.session.User(w, r) // Nil if anonymous or failed to fetch
		if user == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Update last seen in session and in DB in background if necessary
		if time.Since(*user.LastSeen) > 24*time.Hour {

			user.LastSeen = new(time.Now())
			if err := s.session.AddUser(w, r, user); err != nil {
				slog.WarnContext(
					r.Context(),
					"failed to update user last seen in session",
					"error", err,
				)
			}

			go func() {
				// Detach the request context and
				// give this goroutine 5 seconds to finish.
				detachedCtx := context.WithoutCancel(r.Context())
				goCtx, cancel := context.WithTimeout(detachedCtx, 5*time.Second)
				defer cancel()

				if _, err := s.repo.UpdateLastSeen(goCtx, user.ID); err != nil {
					slog.WarnContext(
						r.Context(),
						"failed to update user last seen in DB",
						"error", err,
					)
				}
			}()
		}

		// Try to get the user avatar
		var err error
		user.LocalAvatarURL, err = s.avatar.Get(r.Context(), user)
		if err != nil {
			slog.WarnContext(
				r.Context(),
				"failed to get user avatar",
				"error", err,
			)
		}

		// Put user in context
		ctx := ctxd.WithUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
