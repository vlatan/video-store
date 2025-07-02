package auth

import "factual-docs/internal/models"

var successLogin = models.FlashMessage{
	Message:  "You've been logged in!",
	Category: "info",
}

var failedLogin = models.FlashMessage{
	Message:  "Something went wrong. Login failed!",
	Category: "info",
}

var successLogout = models.FlashMessage{
	Message:  "You've been logged out!",
	Category: "info",
}

var failedLogout = models.FlashMessage{
	Message:  "Something went wrong. Logout failed!",
	Category: "info",
}

var successDeleteAccount = models.FlashMessage{
	Message:  "You accound was deleted!",
	Category: "info",
}

var failedDeleteAccount = models.FlashMessage{
	Message:  "Something went wrong. Account deletion failed!",
	Category: "info",
}
