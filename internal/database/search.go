package database

import (
	"net/url"
	"regexp"
	"strings"
)

var rePunct = regexp.MustCompile(`[^\p{L}\p{N}\s_]+`)
var reSpaces = regexp.MustCompile(`\s+`)

const searchPostsQuery = `
WITH search_terms AS (
    SELECT
        to_tsquery('english', $1) AS and_query,
        to_tsquery('english', $2) AS or_query
),
search_posts AS (
    SELECT
        p.id,
        p.video_id,
        p.title,
        p.thumbnails, (
			SELECT COUNT(*)
			FROM post_like
			WHERE post_like.post_id = p.id
		) AS likes,
		(ts_rank(p.search_vector, st.and_query) * 2) + ts_rank(p.search_vector, st.or_query) AS rank,
        CASE 
            WHEN $4 = 0 THEN COUNT(*) OVER()
            ELSE NULL
        END AS total_results
    FROM post AS p, search_terms AS st
	WHERE p.search_vector @@ st.and_query OR p.search_vector @@ st.or_query
	ORDER BY rank DESC
	LIMIT $3 OFFSET $4
)
SELECT
    video_id,
    title,
    thumbnails,
	likes
FROM search_posts
`

// Get posts based on a search query
func (s *service) SearchPosts(searchQuery string, page int) ([]Post, error) {
	limit := s.config.PostsPerPage
	offset := (page - 1) * limit
	andQuery, orQuery := normalizeSearchQuery(searchQuery)
	return s.queryPosts(searchPostsQuery, andQuery, orQuery, limit, offset)
}

// Remove punctuation and spaces.
// Form two queries where the words are separated with "&" and "|"
func normalizeSearchQuery(input string) (andQuery, orQuery string) {
	// 1. Replace punctuation with spaces
	cleaned := rePunct.ReplaceAllString(input, " ")

	// 2. Collapse multiple spaces into a single space
	cleaned = reSpaces.ReplaceAllString(cleaned, " ")

	// 3. Trim leading/trailing spaces
	cleaned = strings.TrimSpace(cleaned)

	// Handle empty string case after cleaning
	if cleaned == "" {
		return "", ""
	}

	// 4. Split into words
	words := strings.Fields(cleaned) // strings.Fields splits by one or more whitespace characters
	var result []string

	// 5. Discard one letter words
	for _, word := range words {
		if len(word) > 1 {
			result = append(result, word)
		}
	}

	// If there's only one word, both AND and OR queries are the same
	if len(result) == 1 {
		return result[0], result[0]
	}

	andQuery = strings.Join(result, " & ")
	orQuery = strings.Join(result, " | ")

	return andQuery, orQuery
}

// Takes a raw search query and a max length,
// then returns a URL-encoded and truncated string prefixed for Redis.
func EncodeRawSearchQuery(rawQuery string, maxLength int) string {
	// URL-encode the raw query
	encodedQuery := url.QueryEscape(rawQuery)

	// Truncate the URL-encoded query if it exceeds the maximum length
	// Note: We're truncating bytes, which is fine for ASCII/URL-encoded strings.
	// If you were truncating arbitrary UTF-8, you'd need to convert to runes first
	// to avoid splitting multi-byte characters. For URL-encoded strings, this is generally safe.
	if len(encodedQuery) > maxLength {
		encodedQuery = encodedQuery[:maxLength]
	}

	return encodedQuery
}
