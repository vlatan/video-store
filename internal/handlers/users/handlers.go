package users

import (
	"log/slog"
	"net/http"

	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// Handle the user favorites page
func (s *Service) UserFavoritesHandler(w http.ResponseWriter, r *http.Request) {

	// Generate template data
	data := models.GetDataFromContext(r)

	posts, err := s.postsRepo.GetUserFavedPosts(r.Context(), data.CurrentUser.ID, "")

	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get user's favorited posts from DB",
			"path", r.URL.Path,
			"userId", data.CurrentUser.ID,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	data.Posts = &posts
	data.Title = "Your Favorite Documentaries"
	s.ui.RenderHTML(w, r, "user_library.html", data)
}

// Users admin dashboard
func (s *Service) UsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get the page number from the request query param
	page := utils.GetPageNum(r)

	// Generate template data
	data := models.GetDataFromContext(r)

	users, err := s.usersRepo.GetUsers(r.Context(), page)
	if err != nil {
		slog.ErrorContext(
			r.Context(), "failed to get users from DB",
			"path", r.URL.Path,
			"error", err,
		)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if len(users.Items) == 0 {
		http.NotFound(w, r)
		return
	}

	// Assign R2 avatars to users
	for i, user := range users.Items {
		localAvatarURL, err := s.avatars.Get(r.Context(), &user)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "failed to get user's avatar",
				"path", r.URL.Path,
				"userId", user.ID,
				"error", err,
			)
			utils.HttpError(w, http.StatusInternalServerError)
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
