package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Thumbnail struct {
	URL    string `json:"url,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}

type Duration struct {
	ISO   string `json:"iso,omitempty"`
	Human string `json:"human,omitempty"`
}

type Posts struct {
	Items    []Post
	TotalNum int
	TimeTook string
}

type Post struct {
	ID               int        `json:"id,omitempty"`
	VideoID          string     `json:"video_id,omitempty"`
	Title            string     `json:"title,omitempty"`
	Srcset           string     `json:"srcset,omitempty"`
	Thumbnail        *Thumbnail `json:"thumbnail,omitempty"`
	Category         *Category  `json:"category,omitempty"`
	Likes            int        `json:"likes,omitempty"`
	LikeButtonText   string     `json:"like_button_text,omitempty"`
	Description      string     `json:"description,omitempty"`
	ShortDesc        string     `json:"short_description,omitempty"`
	MetaDesc         string     `json:"meta_description,omitempty"`
	RelatedPosts     []Post     `json:"related_posts,omitempty"`
	UploadDate       time.Time  `json:"upload_date,omitempty"`
	Duration         *Duration  `json:"duration,omitempty"`
	CurrentUserLiked bool       `json:"current_user_liked,omitempty"`
	CurrentUserFaved bool       `json:"current_user_faved,omitempty"`
}

var validISO8601 = regexp.MustCompile(`(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?`)

// UPDATE or INSERT a post into the database
func (s *service) UpsertPost(columns ...string) (post Post, err error) {
	return post, err
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
	post.id,
	video_id,
	title, 
	thumbnails, (
		SELECT COUNT(*) FROM post_like
	  	WHERE post_like.post_id = post.id
	) AS likes, 
	description,
 	short_description,
	slug AS category_slug,
	name AS category_name,
	upload_date,
	duration
FROM post 
LEFT JOIN category ON post.category_id = category.id
WHERE video_id = $1 
`

// Get single post from DB based on a video ID
func (s *service) GetSinglePost(videoID string) (post Post, err error) {

	var thumbnails []byte
	var category Category
	var duration Duration

	// Get single row from DB
	err = s.db.QueryRow(getSinglePostQuery, videoID).Scan(
		&post.ID,
		&post.VideoID,
		&post.Title,
		&thumbnails,
		&post.Likes,
		&post.Description,
		&post.ShortDesc,
		&category.Slug,
		&category.Name,
		&post.UploadDate,
		&duration.ISO,
	)

	if err != nil {
		return post, err
	}

	humanDuration, _ := parseISO8601Duration(duration.ISO)
	duration.Human = humanDuration

	post.Category = &category
	post.Duration = &duration

	// Like button text
	post.LikeButtonText = "Like"
	if post.Likes == 1 {
		post.LikeButtonText = "1 Like"
	} else if post.Likes > 1 {
		post.LikeButtonText = fmt.Sprintf("%d Likes", post.Likes)
	}

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
	post.Thumbnail = &maxThumb

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

// Query the DB for posts based on variadic arguments
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
