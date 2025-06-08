package database

import (
	"encoding/json"
	"errors"
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
	VideoID   string    `json:"video_id"`
	Title     string    `json:"title"`
	Srcset    string    `json:"srcset"`
	Thumbnail Thumbnail `json:"thumbnail"`
	Category  Category  `json:"category"`
	Likes     int       `json:"likes"`
	ShortDesc string    `json:"short_description"`
	MetaDesc  string    `json:"meta_description"`
	Related   []Post    `json:"related"`
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

const getSinglePostQuery = `
SELECT 
	title, 
	thumbnails, (
		SELECT COUNT(*) FROM post_like
	  	WHERE post_like.post_id = post.id
	) AS likes, 
 	short_description,
	slug AS category_slug,
	name AS category_name
FROM post 
LEFT JOIN category ON post.category_id = category.id
WHERE video_id = $1 
`

func (s *service) GetSinglePost(videoID string) (post Post, err error) {

	var thumbnails []byte
	var category Category

	// Get single row from DB
	err = s.db.QueryRow(getSinglePostQuery, videoID).Scan(
		&post.Title,
		&thumbnails,
		&post.Likes,
		&post.ShortDesc,
		&category.Slug,
		&category.Name,
	)
	if err != nil {
		return post, err
	}

	// Assign category struct
	post.Category = category

	// Unserialize thumbnails
	thumbsMap, err := unmarshalThumbs(thumbnails)
	if err != nil {
		return post, fmt.Errorf("video ID '%s': %v", videoID, err)
	}

	// Get the thumbnail with the maximum width
	var maxThumb Thumbnail
	for _, thumb := range thumbsMap {
		if thumb.Width > maxThumb.Width {
			maxThumb = thumb
		}
	}

	// Assign the biggest thumbnail to post
	post.Thumbnail = maxThumb

	// Get the first sentence of the short description to be used as meta description
	post.MetaDesc = strings.Split(post.ShortDesc, ".")[0]

	// Make srcset string
	post.Srcset = srcset(thumbsMap, maxThumb.Width)

	return post, err
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
	// Get rows from DB
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

		// Unserialize thumbnails
		thumbsMap, err := unmarshalThumbs(thumbnails)
		if err != nil {
			return posts, fmt.Errorf("video ID '%s': %v", post.VideoID, err)
		}

		post.Srcset = srcset(thumbsMap, 480)
		post.Thumbnail = thumbsMap["medium"]

		// Include the processed post in the result
		posts = append(posts, post)
	}

	// If error during iteration
	if err := rows.Err(); err != nil {
		return posts, err
	}

	return posts, err
}

// Unserialize thumbnails
func unmarshalThumbs(thumbs []byte) (thumbnails map[string]Thumbnail, err error) {
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
