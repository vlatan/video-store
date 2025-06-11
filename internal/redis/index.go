package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Create new redis search index
func (s *service) CreateSearchIndex(ctx context.Context) error {
	schema := []*redis.FieldSchema{
		{FieldName: "video_id", FieldType: redis.SearchFieldTypeText},
		{FieldName: "title", FieldType: redis.SearchFieldTypeText, Weight: 2.0},
		{FieldName: "description", FieldType: redis.SearchFieldTypeText},
		{FieldName: "tags", FieldType: redis.SearchFieldTypeText},
		{FieldName: "thumbnail", FieldType: redis.SearchFieldTypeText},
		{FieldName: "srcset", FieldType: redis.SearchFieldTypeText},
	}
	return s.rdb.FTCreate(ctx, "docs", &redis.FTCreateOptions{}, schema...).Err()
}

func (s *service) UpsertDocument(ctx context.Context, columns ...string) (int, error) {
	post, err := s.db.UpsertPost(columns...)
	if post.ID == 0 || err != nil {
		return 0, err
	}

	// Add to Redis search index
	key := fmt.Sprintf("doc:%d", post.ID)
	err = s.rdb.HSet(ctx, key, map[string]any{
		"video_id":    post.VideoID,
		"title":       post.Title,
		"description": post.ShortDesc,
		// "tags":        post.Tags,
		"thumbnail": post.Thumbnail,
		"srcset":    post.Srcset,
	}).Err()

	return post.ID, err
}
