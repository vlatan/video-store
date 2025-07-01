package auth

import tmpls "factual-docs/internal/services/templates"

var successLogin = tmpls.FlashMessage{
	Message:  "You've been logged in!",
	Category: "info",
}

var failedLogin = tmpls.FlashMessage{
	Message:  "Something went wrong. Login failed!",
	Category: "info",
}

var successLogout = tmpls.FlashMessage{
	Message:  "You've been logged out!",
	Category: "info",
}

var failedLogout = tmpls.FlashMessage{
	Message:  "Something went wrong. Logout failed!",
	Category: "info",
}

var successDeleteAccount = tmpls.FlashMessage{
	Message:  "You accound was deleted!",
	Category: "info",
}

var failedDeleteAccount = tmpls.FlashMessage{
	Message:  "Something went wrong. Account deletion failed!",
	Category: "info",
}
