package posts

import (
	"context"
	"encoding/json"
	"errors"
	"factual-docs/internal/models"
	"factual-docs/internal/shared/database"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Validate video ID
var validVideoID = regexp.MustCompile("^([-a-zA-Z0-9_]{11})$")

// Valid ISO time format
var validISO8601 = regexp.MustCompile(`(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?`)

// Unserialize thumbnails
func unmarshalThumbs(thumbs []byte) (thumbnails map[string]models.Thumbnail, err error) {
	err = json.Unmarshal(thumbs, &thumbnails)
	if err != nil {
		return thumbnails, err
	}

	// Check if no thumbnails at all
	if len(thumbnails) == 0 {
		return thumbnails, errors.New("no thumbnails found")
	}

	return thumbnails, err
}

// Parse ISO8601 duration in a human readable string
func parseISO8601Duration(duration string) (string, error) {
	// Remove PT prefix
	if !strings.HasPrefix(duration, "PT") {
		return "", fmt.Errorf("invalid duration format: %s", duration)
	}
	duration = strings.TrimPrefix(duration, "PT")

	// Find the substrings (hours, minutes, seconds)
	matches := validISO8601.FindStringSubmatch(duration)
	if len(matches) == 0 {
		return "", fmt.Errorf("invalid duration format: %s", duration)
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	sec, _ := strconv.ParseFloat(matches[3], 64)
	seconds := int(sec)

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds), nil
}

// Check if category is valid
func isValidCategory(categories []database.Category, slug string) (database.Category, bool) {
	for _, category := range categories {
		if category.Slug == slug {
			return category, true
		}
	}
	return database.Category{}, false
}

// Get post's related posts based on provided title as search query
func (s *Service) getRelatedPosts(ctx context.Context, title string) (posts []models.Post, err error) {
	// Search the DB for posts
	searchedPosts, err := s.SearchPosts(ctx, title, s.config.NumRelatedPosts+1, 0)

	if err != nil {
		return posts, err
	}

	for _, sp := range searchedPosts.Items {
		if sp.Title != title {
			posts = append(posts, sp)
		}
	}

	return posts, err
}

// Create a srcset string from a map of thumbnails
func srcset(thumbnails map[string]models.Thumbnail, maxWidth int) string {

	// Get the Thumbnail structs from the map
	items := make([]models.Thumbnail, 0, len(thumbnails))
	for _, item := range thumbnails {
		items = append(items, item)
	}

	// Sort the thumbnails by width
	sort.Slice(items, func(i, j int) bool {
		return items[i].Width < items[j].Width
	})

	// Create the srcset string
	var result string
	for _, item := range items {
		if item.Width <= maxWidth {
			result += fmt.Sprintf("%s %dw, ", item.URL, item.Width)
		}
	}

	return strings.TrimSuffix(result, ", ")
}

// Query the DB for posts based on variadic arguments
func (s *Service) queryPosts(
	ctx context.Context,
	query string,
	args ...any,
) (posts []models.Post, err error) {
	// Get rows from DB
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return posts, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post models.Post
		var thumbnails []byte

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&post.VideoID, &post.Title, &thumbnails, &post.Likes); err != nil {
			return posts, err
		}

		// Unserialize thumbnails
		thumbsMap, err := unmarshalThumbs(thumbnails)
		if err != nil {
			return posts, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
		}

		// Craft srcset string
		post.Srcset = srcset(thumbsMap, 480)
		thumb := thumbsMap["medium"]
		post.Thumbnail = &thumb

		// Include the processed post in the result
		posts = append(posts, post)
	}

	// If error during iteration
	if err := rows.Err(); err != nil {
		return posts, err
	}

	return posts, err
}
