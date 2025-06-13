package database

import (
	"fmt"
)

const searchPostsQuery = `
WITH search_terms AS (
	SELECT
		lexeme AS and_query,
		to_tsquery('english', replace(lexeme::text, ' & ', ' | ')) AS or_query,
		replace(lexeme::text, ' & ', ' ') AS raw_query
	FROM plainto_tsquery('english', $1) AS lexeme
)
SELECT
    p.video_id,
    p.title,
    p.thumbnails,
    (
        SELECT COUNT(*)
        FROM post_like
        WHERE post_like.post_id = p.id
    ) AS likes,
    CASE 
        WHEN $3 = 0 THEN COUNT(*) OVER()
        ELSE 0
    END AS total_results
FROM post AS p, search_terms AS st
WHERE p.search_vector @@ st.and_query OR p.search_vector @@ st.or_query
ORDER BY 
	(ts_rank(p.search_vector, st.and_query, 32) * 2) + 
	ts_rank(p.search_vector, st.or_query, 32) +
	(similarity(p.title, st.raw_query) * 0.5) DESC,
	likes DESC,
	p.upload_date DESC
LIMIT $2 OFFSET $3
`

// Get posts based on a user search query
// Transform the user query into two queries with words separated by '&' and '|'
func (s *service) SearchPosts(searchTerm string, limit, offset int) (posts Posts, err error) {

	// andQuery, orQuery := normalizeSearchQuery(searchQuery)

	// Get rows from DB
	rows, err := s.db.Query(searchPostsQuery, searchTerm, limit, offset)
	if err != nil {
		return posts, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var post Post
		var thumbnails []byte
		var totalNum int

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&post.VideoID, &post.Title, &thumbnails, &post.Likes, &totalNum); err != nil {
			return posts, err
		}

		// Unserialize thumbnails
		thumbsMap, err := unmarshalThumbs(thumbnails)
		if err != nil {
			return posts, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
		}

		post.Srcset = srcset(thumbsMap, 480)
		thumb := thumbsMap["medium"]
		post.Thumbnail = &thumb

		// Include the processed post in the result
		posts.Items = append(posts.Items, post)
		if totalNum != 0 {
			posts.TotalNum = totalNum
		}
	}

	// If error during iteration
	if err := rows.Err(); err != nil {
		return posts, err
	}

	return posts, err
}
