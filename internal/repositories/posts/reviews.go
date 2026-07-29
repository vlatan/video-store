package posts

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/vlatan/video-store/internal/models"
	"github.com/vlatan/video-store/internal/utils"
)

// GetPostReviews gets posts reviews
func (r *Repository) GetPostReviews(ctx context.Context, videoID, cursor string) (models.Reviews, error) {

	// The video ID and the limit are the first two arguments ($1 and $2)
	// Peek for one review beoynd the limit to see if there are more reviews,
	// meaning whether to construct and send the next cursor at all.
	args := []any{videoID, r.config.ReviewsPerPost + 1}

	// The default template variables - SQL parts
	var andCondition string
	total := "COUNT(*) OVER()"

	var zero, reviews models.Reviews

	// Build args and SQL parts.
	// If cursor (meanining this is not the first page),
	// do not count total and supply AND clause
	if cursor != "" {

		total = "0"
		cursorParts, err := decodeCursor(cursor)
		if err != nil {
			return zero, err
		}

		if len(cursorParts) != 2 {
			return zero, errors.New("invalid cursor components")
		}

		args = append(args, cursorParts[1], cursorParts[2])
		andCondition = "AND (prev.created_at, prev.rating_id) < ($3, $4)"
	}

	sqlParts := struct{ TotalCount, AndCondition string }{total, andCondition}
	query, err := r.GetQuery("post_reviews.sql", sqlParts)
	if err != nil {
		return zero, err
	}

	// Get rows from DB
	rows, err := r.db.Pool.Query(ctx, query, args...)
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
			review   models.Review
			totalNum int
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
			&totalNum,
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

		// Assign the total num of reviews
		reviews.TotalNum = max(reviews.TotalNum, totalNum)

		// This review is done, append it
		reviews.Items = append(reviews.Items, review)
	}

	// If error during iteration
	if err = rows.Err(); err != nil {
		return zero, err
	}

	return reviews, nil
}
