package models

import (
	"context"
)

type contextKey struct {
	name string
}

// Universal context key to get the user from context
var UserContextKey = contextKey{name: "user"}

// Universal context key to get the page data from context
var DataContextKey = contextKey{name: "data"}

// GetUserFromContext gets the user from context
func GetUserFromContext(ctx context.Context) *User {
	user, _ := ctx.Value(UserContextKey).(*User)
	return user // nil if user not in context
}
