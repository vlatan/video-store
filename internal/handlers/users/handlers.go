package users

import "net/http"

// Handle the Home page
func (s *Service) UserLibraryHandler(w http.ResponseWriter, r *http.Request) {
	// Generate template data
	data := s.tm.NewData(w, r)
	data.CurrentUser = s.auth.GetUserFromContext(r)

}
