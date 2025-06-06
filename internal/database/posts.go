package database

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Thumbnail struct {
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Post struct {
	VideoID      string    `json:"video_id"`
	Title        string    `json:"title"`
	Srcset       string    `json:"srcset"`
	Thumbnail    Thumbnail `json:"thumbnail"`
	CategorySlug string    `json:"category_slug"`
	CategoryName string    `json:"category_name"`
	Likes        int       `json:"likes"`
}

const getPostsQuery = `
SELECT video_id, title, thumbnails, (
	SELECT COUNT(*) FROM post_like
	WHERE post_like.post_id = post.id
) AS likes FROM post
ORDER BY %s
LIMIT $1 OFFSET $2
`

// Get a limited number of posts with offset
func (s *service) GetPosts(page int, orderBy string) ([]Post, error) {

	limit := s.config.PostsPerPage
	offset := (page - 1) * limit

	order := "upload_date DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	query := fmt.Sprintf(getPostsQuery, order)
	return s.queryPosts(query, limit, offset)
}

const getCategoryPostsQuery = `
SELECT video_id, title, thumbnails, (
	SELECT COUNT(*) FROM post_like
	WHERE post_like.post_id = post.id
) AS likes FROM post 
WHERE category_id = (SELECT id FROM category WHERE slug = $1) 
ORDER BY %s
LIMIT $2 OFFSET $3
`

// Get a limited number of posts from one category with offset
func (s *service) GetCategoryPosts(categorySlug, orderBy string, page int) ([]Post, error) {

	limit := s.config.PostsPerPage
	offset := (page - 1) * limit

	order := "upload_date DESC"
	if orderBy == "likes" {
		order = "likes DESC, " + order
	}

	query := fmt.Sprintf(getCategoryPostsQuery, order)
	return s.queryPosts(query, categorySlug, limit, offset)
}

// Convert DB post to e ready post for templates
func processPost(post *Post, thumbs []byte) error {
	// Unmarshall the byte thumbnails into a map of structs
	var thumbnails map[string]Thumbnail
	err := json.Unmarshal(thumbs, &thumbnails)
	if err != nil {
		return err
	}

	post.Srcset = srcset(thumbnails, 480)
	post.Thumbnail = thumbnails["medium"]

	return nil

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
		var post Post
		var thumbnails []byte

		// Paste post from row to struct, thumbnails in a separate var
		if err = rows.Scan(&post.VideoID, &post.Title, &thumbnails, &post.Likes); err != nil {
			return posts, err
		}

		// Process the post
		err = processPost(&post, thumbnails)
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
