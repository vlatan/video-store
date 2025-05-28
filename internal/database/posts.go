package database

import (
	"encoding/json"
)

type Post struct {
	VideoID, Title string
	Thumbnails     []byte
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ProcessedPost struct {
	VideoID, Title, Srcset string
	Thumbnail              Thumbnail
}

const getPostsQuery = `
SELECT video_id, title, thumbnails_json FROM post 
ORDER BY upload_date DESC
LIMIT $1 OFFSET $2
`

// Get a limited number of posts with offset
func (s *service) GetPosts(page int) ([]ProcessedPost, error) {

	limit := s.config.PostsPerPage
	offset := page * limit

	rows, err := s.db.Query(getPostsQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		posts          []Post
		processedPosts []ProcessedPost
	)

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.VideoID, &post.Title, &post.Thumbnails); err != nil {
			return processedPosts, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return processedPosts, err
	}

	for _, post := range posts {
		var pPost ProcessedPost

		pPost.VideoID = post.VideoID
		pPost.Title = post.Title

		var thumbnails map[string]Thumbnail
		err := json.Unmarshal(post.Thumbnails, &thumbnails)
		if err != nil {
			return processedPosts, err
		}

		pPost.Thumbnail = thumbnails["medium"]
		processedPosts = append(processedPosts, pPost)
	}

	return processedPosts, nil
}
