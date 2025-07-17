package users

import (
	"factual-docs/internal/shared/utils"
	"log"
	"net/http"
	"time"
)

// Handle the user favorites page
func (s *Service) UserFavoritesHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page number from the request query param
	page := utils.GetPageNum(r)

	// Generate template data
	data := s.ui.NewData(w, r)

	posts, err := s.postsRepo.GetUserFavedPosts(r.Context(), data.CurrentUser.ID, page)

	if err != nil {
		log.Printf("Was unabale to fetch posts on URI '%s': %v", r.RequestURI, err)
		if page > 1 {
			s.ui.JSONError(w, r, http.StatusInternalServerError)
			return
		}
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	if page > 1 && len(posts.Items) == 0 {
		log.Printf("Fetched zero posts on URI '%s'", r.RequestURI)
		s.ui.JSONError(w, r, http.StatusNotFound)
		return
	}

	// If not the first page return JSON
	if page > 1 {
		time.Sleep(time.Millisecond * 400)
		s.ui.WriteJSON(w, r, posts.Items)
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
	data := s.ui.NewData(w, r)

	users, err := s.usersRepo.GetUsers(r.Context(), page)
	if err != nil {
		log.Printf("Was unabale to fetch users on URI '%s': %v", r.RequestURI, err)
		s.ui.HTMLError(w, r, http.StatusInternalServerError, data)
		return
	}

	avatars := s.GetAvatars(r.Context(), users.Items)
	for avatar := range avatars {
		users.Items[avatar.index].LocalAvatarURL = avatar.localAvatar
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
