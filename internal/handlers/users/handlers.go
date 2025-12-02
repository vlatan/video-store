package users

import (
	"log"
	"net/http"

	"github.com/vlatan/video-store/internal/utils"
)

// Handle the user favorites page
func (s *Service) UserFavoritesHandler(w http.ResponseWriter, r *http.Request) {

	// Get the cursor if any
	cursor := r.URL.Query().Get("cursor")

	// Generate template data
	data := utils.GetDataFromContext(r)

	posts, err := s.postsRepo.GetUserFavedPosts(r.Context(), data.CurrentUser.ID, cursor)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	// If there's a cursor this is not the first page, return JSON
	if cursor != "" {
		s.ui.WriteJSON(w, r, posts)
		return
	}

	data.Posts = posts
	data.Title = "Your Favorite Documentaries:"
	s.ui.RenderHTML(w, r, "user_library.html", data)
}

// Users admin dashboard
func (s *Service) UsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get the page number from the request query param
	page := utils.GetPageNum(r)

	// Generate template data
	data := utils.GetDataFromContext(r)

	users, err := s.usersRepo.GetUsers(r.Context(), page)
	if err != nil {
		log.Printf("was unabale to fetch users on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	if len(users.Items) == 0 {
		http.NotFound(w, r)
		return
	}

	// Assign local avatars to users
	if err = s.SetAvatars(r.Context(), users.Items); err != nil {
		log.Printf("was unabale to set users avatars on URI '%s': %v", r.RequestURI, err)
		utils.HttpError(w, http.StatusInternalServerError)
		return
	}

	data.PaginationInfo = s.ui.NewPagination(
		page,
		users.TotalNum,
		s.config.PostsPerPage,
	)

	data.Users = users
	data.Title = "Admin Dashboard"
	s.ui.RenderHTML(w, r, "admin.html", data)
}
