package auth

import (
	"github.com/vlatan/video-store/internal/drivers/store"
)

var successLogin = store.FlashMessage{
	Message:  "You've been logged in!",
	Category: "info",
}

var failedLogin = store.FlashMessage{
	Message:  "Something went wrong. Login failed!",
	Category: "info",
}

var successLogout = store.FlashMessage{
	Message:  "You've been logged out!",
	Category: "info",
}

var failedLogout = store.FlashMessage{
	Message:  "Something went wrong. Logout failed!",
	Category: "info",
}

var successDeleteAccount = store.FlashMessage{
	Message:  "You accound was deleted!",
	Category: "info",
}

var failedDeleteAccount = store.FlashMessage{
	Message:  "Something went wrong. Account deletion failed!",
	Category: "info",
}
