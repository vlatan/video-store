package user

import (
	"log/slog"
	"net/http"
)

// Handle the user favorites page
func (s *Service) Favorites(w http.ResponseWriter, r *http.Request) {

	// Get template data
	data := s.tmpl.Data(w, r)

	posts, err := s.Repo.Favorites(r.Context(), data.CurrentUser.ID, "")

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get user fav posts from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	data.Posts = &posts
	data.Title = "Your Favorite Documentaries"
	s.tmpl.RenderHTML(w, r, "user_library.html", data)
}

// Users admin dashboard
func (s *Service) All(w http.ResponseWriter, r *http.Request) {
	// Get the page number from the request query param
	page := getPageNum(r)

	// Get template data
	data := s.tmpl.Data(w, r)

	users, err := s.Repo.All(r.Context(), page)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get users from DB",
			"error", err,
		)
		s.tmpl.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	if len(users.Items) == 0 {
		slog.WarnContext(r.Context(), "no users found in DB")
		s.tmpl.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Assign R2 avatars to users
	for i, user := range users.Items {
		localAvatarURL, err := s.avatar.Get(r.Context(), &user)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to get user avatar",
				"error", err,
			)
			s.tmpl.HTMLError(w, r, data, http.StatusInternalServerError)
			return
		}
		users.Items[i].LocalAvatarURL = localAvatarURL
	}

	data.PaginationInfo = s.tmpl.NewPagination(
		page,
		users.TotalNum,
		s.cfg.PostsPerPage,
	)

	data.Users = &users
	data.Title = "Users"
	s.tmpl.RenderHTML(w, r, "admin.html", data)
}
