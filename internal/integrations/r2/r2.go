package r2

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/vlatan/video-store/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type Service interface {
	// DeleteObject removes an object from bucket
	DeleteObject(ctx context.Context, bucket, key string) error
	// ObjectExists checks if the object exists in the bucket
	ObjectExists(ctx context.Context, timeout time.Duration, bucket, key string) error
	// HeadObject gets and returns the head of a given object
	HeadObject(ctx context.Context, bucket, key string) (*s3.HeadObjectOutput, error)

	// PutObject puts object to bucket having the content
	PutObject(
		ctx context.Context,
		bucket string,
		key string,
		body io.Reader,
		contentType string,
		metadata map[string]string,
	) error

	// UploadFile uploads a file to bucket
	UploadFile(ctx context.Context, bucket, rootPath, key, filePath string) error
}

type service struct {
	client *s3.Client
}

// New creates a new R2 client
func New(ctx context.Context, cfg *config.Config) Service {

	// Create SDK config for an R2 service
	// An ordinary AWS SDK config would look like:
	// sdkConfig, err := awsConfig.LoadDefaultConfig(ctx)
	sdkConfig, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.R2AccessKeyId, cfg.R2SecretAccessKey, ""),
		),
		awsConfig.WithRegion("auto"),
	)

	if err != nil {
		log.Fatalf("failed to load AWS/R2 SDK configuration, %v", err)
	}

	// Create the R2 client
	// An ordinary AWS client would look like:
	// client := s3.NewFromConfig(sdkConfig)
	client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		baseEndpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountId)
		o.BaseEndpoint = aws.String(baseEndpoint)
	})

	return &service{client}
}

// DeleteObject removes an object from bucket
func (s *service) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	return err
}

// ObjectExists checks if the object exists in the bucket
func (s *service) ObjectExists(ctx context.Context, timeout time.Duration, bucket, key string) error {
	return s3.NewObjectExistsWaiter(s.client).Wait(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		},
		timeout,
	)
}

// HeadObject gets and returns the head of a given object
func (s *service) HeadObject(ctx context.Context, bucket, key string) (*s3.HeadObjectOutput, error) {
	return s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

// PutObject puts object to bucket having the content
func (s *service) PutObject(
	ctx context.Context,
	bucket string,
	key string,
	body io.Reader,
	contentType string,
	metadata map[string]string,
) error {

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
		Metadata:    metadata,
	})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			return fmt.Errorf(
				"error while uploading object to %s; The object is too large: %w",
				bucket, err,
			)

		}

		return fmt.Errorf(
			"couldn't upload object %s:%s: %w",
			bucket, key, err,
		)
	}

	if err = s.ObjectExists(ctx, time.Minute, bucket, key); err != nil {
		return fmt.Errorf(
			"failed attempt to wait for object %s:%s to exist: %w",
			bucket, key, err,
		)
	}

	return nil
}

// UploadFile uploads a file to bucket
func (s *service) UploadFile(ctx context.Context, bucket, rootPath, key, filePath string) error {

	file, err := SecureOpen(rootPath, filePath)
	if err != nil {
		return fmt.Errorf("couldn't open the file %s: %w", filePath, err)
	}
	defer file.Close()

	// Read the first 512 bytes for content type detection
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) { // EOF is expected if file is smaller than 512 bytes
		return fmt.Errorf("couldn't read the file %s: %w", filePath, err)
	}

	// Seek back to the beginning for the actual upload
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return fmt.Errorf("couldn't seek to beginning of file %s: %w", filePath, err)
	}

	contentType := http.DetectContentType(buffer)
	return s.PutObject(ctx, bucket, key, file, contentType, nil)
}
