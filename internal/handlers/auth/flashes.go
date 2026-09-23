package auth

import "github.com/vlatan/video-store/internal/types"

var successLogin = types.FlashMessage{
	Message:  "You've been logged in!",
	Category: "info",
}

var failedLogin = types.FlashMessage{
	Message:  "Something went wrong. Login failed!",
	Category: "info",
}

var successLogout = types.FlashMessage{
	Message:  "You've been logged out!",
	Category: "info",
}

var failedLogout = types.FlashMessage{
	Message:  "Something went wrong. Logout failed!",
	Category: "info",
}

var successDeleteAccount = types.FlashMessage{
	Message:  "You accound was deleted!",
	Category: "info",
}

var failedDeleteAccount = types.FlashMessage{
	Message:  "Something went wrong. Account deletion failed!",
	Category: "info",
}
