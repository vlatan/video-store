package server

import "factual-docs/internal/templates"

var successLogin = templates.FlashMessage{
	Message:  "You've been logged in!",
	Category: "info",
}

var failedLogin = templates.FlashMessage{
	Message:  "Something went wrong. Login failed!",
	Category: "info",
}

var successLogout = templates.FlashMessage{
	Message:  "You've been logged out!",
	Category: "info",
}

var failedLogout = templates.FlashMessage{
	Message:  "Something went wrong. Logout failed",
	Category: "info",
}

var successDeleteAccount = templates.FlashMessage{
	Message:  "You accound was deleted!",
	Category: "info",
}

var failedDeleteAccount = templates.FlashMessage{
	Message:  "Something went wrong. Account deletion failed",
	Category: "info",
}
