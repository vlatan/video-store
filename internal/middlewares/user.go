package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/ctxd"
	"github.com/vlatan/video-store/internal/types"
	"github.com/vlatan/video-store/internal/utils/pathx"
)

// LoadUser gets the user data from session and DB and stores it in the context
func (s *Service) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Skip static files, they don't need user
		if pathx.IsStatic(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Fetch the user ID and tokens from session
		sessUser := s.session.User(w, r) // Nil if anonymous or failed to fetch
		if sessUser == nil {
			next.ServeHTTP(w, r)
			return
		}

		// Put user ID in context
		ctx := ctxd.WithUserID(r.Context(), sessUser.ID)

		// Put user loader in context
		ctx = ctxd.WithUserLoader(ctx, &types.UserLoader{

			// Define the Load function.
			// Enclose the logic necessary to get the user.
			Load: func() (*types.User, error) {

				// Fetch the user data from DB
				user, err := s.usersRepo.GetSingleUser(r.Context(), sessUser.ID)
				if err != nil {
					return nil, err
				}

				// Attach tokens and config to user object
				user.AccessToken = sessUser.AccessToken
				user.RefreshToken = sessUser.RefreshToken
				user.Expiry = sessUser.Expiry
				user.Config = s.config

				// Update last seen in DB if necessary
				if time.Since(*user.LastSeen) > 24*time.Hour {
					go func() {
						// Detach the request context and
						// give this goroutine 5 seconds to finish.
						detachedCtx := context.WithoutCancel(r.Context())
						goroutineCtx, cancel := context.WithTimeout(detachedCtx, 5*time.Second)
						defer cancel()

						_, err := s.usersRepo.UpdateLastSeen(goroutineCtx, user.ID)
						if err != nil {
							slog.WarnContext(
								r.Context(),
								"failed to update user last seen in DB",
								"error", err,
							)
						}
					}()
				}

				// Try to get the user avatar
				user.LocalAvatarURL, err = s.avatar.Get(r.Context(), &user)
				if err != nil {
					slog.WarnContext(
						r.Context(),
						"failed to get user avatar",
						"error", err,
					)
				}

				return &user, nil
			},
		})

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
