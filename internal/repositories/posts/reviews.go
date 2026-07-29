package posts

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// GetPostReviews gets posts reviews
func (r *Repository) GetPostReviews(ctx context.Context, videoID, cursor string) (models.Reviews, error) {

	var zero, reviews models.Reviews

	query, err := r.GetQuery("post_reviews.sql", nil)
	if err != nil {
		return zero, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, videoID)
	if err != nil {
		return zero, err
	}

	// Close rows on exit
	defer rows.Close()

	// Define policies for review headline and content sanitization
	strictPolicy := bluemonday.StrictPolicy()
	simplePolicy := utils.SimplePolicy()

	// Iterate over the rows
	for rows.Next() {

		var (
			review models.Review
			name,
			email,
			avatarURL,
			analyticsID sql.NullString
		)

		err = rows.Scan(
			&review.User.ID,
			&review.User.ProviderUserId,
			&review.User.Provider,
			&name,
			&email,
			&avatarURL,
			&analyticsID,
			&review.Headline,
			&review.Content,
			&review.Rating,
			&review.UpdatedAt,
		)

		if err != nil {
			return zero, err
		}

		review.User.Name = name.String
		review.User.Email = email.String
		review.User.AvatarURL = avatarURL.String
		review.User.AnalyticsID = analyticsID.String

		// Sanitize headline
		safeHeadline := strictPolicy.Sanitize(review.Headline)
		review.HTMLHeadline = template.HTML(safeHeadline) // #nosec G203

		// Convert to HTML and sanitize content
		safeHTMLcontent, err := utils.ParseMarkdown(review.Content, simplePolicy)
		if err != nil {
			return zero, fmt.Errorf(
				"could not parse/sanitize markdown on video %q review: %v",
				videoID, err,
			)
		}
		review.HTMLContent = template.HTML(safeHTMLcontent) // #nosec G203

		// This review is done, append it
		reviews.Items = append(reviews.Items, review)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return zero, err
	}

	return reviews, nil
}
