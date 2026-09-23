package users

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils/ctxv"
)

// Handle the user favorites page
func (s *Service) UserFavoritesHandler(w http.ResponseWriter, r *http.Request) {

	// Generate template data
	data := ctxv.Get[*models.TemplateData](r.Context())

	posts, err := s.postsRepo.GetUserFavedPosts(r.Context(), data.CurrentUser.ID, "")

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get user fav posts from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	data.Posts = &posts
	data.Title = "Your Favorite Documentaries"
	s.ui.RenderHTML(w, r, "user_library.html", data)
}

// Users admin dashboard
func (s *Service) UsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get the page number from the request query param
	page := GetPageNum(r)

	// Get template data
	data := ctxv.Get[*models.TemplateData](r.Context())

	users, err := s.usersRepo.GetUsers(r.Context(), page)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get users from DB",
			"error", err,
		)
		s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
		return
	}

	if len(users.Items) == 0 {
		slog.WarnContext(r.Context(), "no users found in DB")
		s.ui.HTMLError(w, r, data, http.StatusNotFound)
		return
	}

	// Assign R2 avatars to users
	for i, user := range users.Items {
		localAvatarURL, err := s.avatars.Get(r.Context(), &user)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to get user avatar",
				"error", err,
			)
			s.ui.HTMLError(w, r, data, http.StatusInternalServerError)
			return
		}
		users.Items[i].LocalAvatarURL = localAvatarURL
	}

	data.PaginationInfo = s.ui.NewPagination(
		page,
		users.TotalNum,
		s.config.PostsPerPage,
	)

	data.Users = &users
	data.Title = "Users"
	s.ui.RenderHTML(w, r, "admin.html", data)
}
