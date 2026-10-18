package handler

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/ctxd"
)

// Handle the user favorites page
func (s *Service) Favorites(w http.ResponseWriter, r *http.Request) {

	currentUser := ctxd.GetUser(r.Context())
	posts, err := s.post.Favorites(r.Context(), currentUser.ID, "")

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get user fav posts from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	// Get template data
	data := userData{
		Title: "Your Favorite Documentaries",
		Posts: posts,
		Base:  s.tmpl.Base(w, r),
	}

	s.tmpl.RenderHTML(w, r, "user_library.html", data)
}

// Users admin dashboard
func (s *Service) All(w http.ResponseWriter, r *http.Request) {
	// Get the page number from the request query param
	page := getPageNum(r)

	users, err := s.user.All(r.Context(), page)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get users from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
		return
	}

	if len(users.Items) == 0 {
		slog.WarnContext(r.Context(), "no users found in DB")
		s.tmpl.HTMLError(w, r, http.StatusNotFound)
		return
	}

	// Assign R2 avatars to users
	for i, user := range users.Items {
		localAvatarURL, err := s.avatar.Get(r.Context(), user.PublicID, user.AvatarURL)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to get user avatar",
				"error", err,
			)
			s.tmpl.HTMLError(w, r, http.StatusInternalServerError)
			return
		}
		users.Items[i].LocalAvatarURL = localAvatarURL
	}

	// Get template data
	data := userData{
		Title: "Users",
		Users: users,
		PaginationInfo: s.tmpl.NewPagination(
			page,
			users.TotalNum,
			s.cfg.PostsPerPage,
		),
		Base: s.tmpl.Base(w, r)}

	s.tmpl.RenderHTML(w, r, "admin.html", data)
}
