package models

import "net/http"

type contextKey struct {
	name string
}

// Universal context key to get the user from context
var UserContextKey = contextKey{name: "user"}

// Universal context key to get the page data from context
var DataContextKey = contextKey{name: "data"}

// GetUserFromContext gets the user from context
func GetUserFromContext(r *http.Request) *User {
	user, _ := r.Context().Value(UserContextKey).(*User)
	return user // nil if user not in context
}

// GetDataFromContext gets the template data from context
func GetDataFromContext(r *http.Request) *TemplateData {
	data, _ := r.Context().Value(DataContextKey).(*TemplateData)
	return data // nil if data not in context
}
