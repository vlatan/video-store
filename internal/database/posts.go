package database

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Post struct {
	VideoID    string `db:"video_id"`
	Title      string `db:"title"`
	Thumbnails []byte `db:"title"`
}

const getPostsQuery = `
SELECT video_id, title, thumbnails FROM post 
ORDER BY upload_date DESC
LIMIT $1 OFFSET $2
`

// Get a limited number of posts with offset
func (s *service) GetPosts(page int) ([]Post, error) {

	limit := s.config.PostsPerPage
	offset := page * limit

	rows, err := s.db.Query(getPostsQuery, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.VideoID, &post.Title, &post.Thumbnails); err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return posts, err
	}

	return posts, nil
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type PPost struct {
	VideoID   string    `json:"video_id"`
	Title     string    `json:"title"`
	Srcset    string    `json:"srcset"`
	Thumbnail Thumbnail `json:"thumbnail"`
}

// Get posts where thumbnails are processed
func (s *service) GetProcessedPosts(page int) ([]PPost, error) {

	var pPosts []PPost
	posts, err := s.GetPosts(page)
	if err != nil {
		return pPosts, err
	}

	for _, post := range posts {
		pPost := PPost{
			VideoID: post.VideoID,
			Title:   post.Title,
		}

		var thumbnails map[string]Thumbnail
		err := json.Unmarshal(post.Thumbnails, &thumbnails)
		if err != nil {
			return pPosts, err
		}

		pPost.Srcset = srcset(thumbnails, 480)
		pPost.Thumbnail = thumbnails["medium"]
		pPosts = append(pPosts, pPost)

	}

	return pPosts, nil
}

// Create a srcset string from a map of thumbnails
func srcset(thumbnails map[string]Thumbnail, maxWidth int) string {

	// Get the Thumbnail structs from the map
	items := make([]Thumbnail, 0, len(thumbnails))
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
