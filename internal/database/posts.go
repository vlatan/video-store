package database

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type DBPost struct {
	VideoID    string `db:"video_id"`
	Title      string `db:"title"`
	Thumbnails []byte `db:"title"`
}

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Post struct {
	VideoID   string    `json:"video_id"`
	Title     string    `json:"title"`
	Srcset    string    `json:"srcset"`
	Thumbnail Thumbnail `json:"thumbnail"`
}

const getPostsQuery = `
SELECT video_id, title, thumbnails FROM post 
ORDER BY upload_date DESC
LIMIT $1 OFFSET $2
`

// Get a limited number of posts with offset
func (s *service) GetPosts(page int) ([]Post, error) {

	limit := s.config.PostsPerPage
	offset := (page - 1) * limit

	return s.queryPosts(getPostsQuery, limit, offset)
}

const getCategoryPostsQuery = `
SELECT video_id, title, thumbnails FROM post 
WHERE category_id = (SELECT id FROM category WHERE slug = $1) 
ORDER BY upload_date DESC 
LIMIT $2 OFFSET $3
`

// Get a limited number of posts from one category with offset
func (s *service) GetCategoryPosts(categorySlug string, page int) ([]Post, error) {

	limit := s.config.PostsPerPage
	offset := (page - 1) * limit

	return s.queryPosts(getCategoryPostsQuery, categorySlug, limit, offset)
}

// Convert DB post to e ready post for templates
func processPost(dbPost DBPost) (Post, error) {
	// Unmarshall the thumbnails into a map of structs
	var thumbnails map[string]Thumbnail
	err := json.Unmarshal(dbPost.Thumbnails, &thumbnails)
	if err != nil {
		return Post{}, err
	}

	// Construct the processed post
	post := Post{
		VideoID:   dbPost.VideoID,
		Title:     dbPost.Title,
		Srcset:    srcset(thumbnails, 480),
		Thumbnail: thumbnails["medium"],
	}

	return post, nil

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

func (s *service) queryPosts(query string, args ...any) (posts []Post, err error) {
	// Query the rows
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return posts, err
	}

	// Close rows on exit
	defer rows.Close()

	// Iterate over the rows
	for rows.Next() {
		var dbPost DBPost

		// Paste post from row to struct
		if err = rows.Scan(&dbPost.VideoID, &dbPost.Title, &dbPost.Thumbnails); err != nil {
			return posts, err
		}

		// Process the post
		post, err := processPost(dbPost)
		if err != nil {
			return posts, err
		}

		// Include the processed post in the result
		posts = append(posts, post)
	}

	// If error during iteration
	if err := rows.Err(); err != nil {
		return posts, err
	}

	return posts, err
}
