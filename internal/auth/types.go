package auth

type contextKey struct {
	name string
}

var UserContextKey = contextKey{name: "user"}
