package posts

import "factual-docs/internal/models"

var failedForm = models.FlashMessage{
	Message:  "Something went wrong!",
	Category: "info",
}
