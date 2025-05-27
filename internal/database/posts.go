package database

type Post struct {
	VideoID    string
	Title      string
	Thumbnails []byte
}

const getPostsQuery = `
SELECT video_id, title, thumbnails FROM post 
ORDER BY upload_date DESC
LIMIT $1 OFFSET $2
`

// Get a limitet number of posts with offset
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
