package redis

import (
	"context"
	"factual-docs/internal/database"
	"fmt"
	"log"

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

func (s *service) DeleteDocument(ctx context.Context, id int) error {
	// TODO: delete post method

	// Remove from Redis
	key := fmt.Sprintf("doc:%d", id)
	return s.rdb.Del(ctx, key).Err()
}

func (s *service) SearchDocuments(ctx context.Context, query string) (posts []database.Post, err error) {
	result, err := s.rdb.FTSearch(ctx, "docs", query).Result()
	if err != nil {
		return nil, err
	}

	// TODO: Convert the doc to Post object
	for _, doc := range result.Docs {
		log.Println(doc.Fields)
		// posts = append(posts, doc.Fields)
	}

	return posts, nil
}

func (s *service) SyncExistingData(ctx context.Context) error {
	// TODO: Clean way to sync the D and th index
	return nil
}
