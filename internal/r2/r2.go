package r2

import (
	"context"
	"errors"
	"factual-docs/internal/config"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type Service interface {
	// ObjectExists checks if the object exists in the bucket
	ObjectExists(ctx context.Context, bucketName, objectKey string) error
	// PutObject puts object to bucket having the content
	PutObject(ctx context.Context, body io.Reader, contentType, bucketName, objectKey string) error
	// UploadFile uploads a file to bucket
	UploadFile(ctx context.Context, bucketName, objectKey, fileName string) error
}

type service struct {
	client *s3.Client
}

var (
	r2Instance *service
	once       sync.Once
)

// New creates a new R2 client
func New(ctx context.Context, cfg *config.Config) Service {

	once.Do(func() {
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

		r2Instance = &service{
			client: client,
		}
	})

	return r2Instance
}

// ObjectExists checks if the object exists in the bucket
func (s *service) ObjectExists(ctx context.Context, bucketName, objectKey string) error {
	return s3.NewObjectExistsWaiter(s.client).Wait(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		},
		time.Minute,
	)
}

// PutObject puts object to bucket having the content
func (s *service) PutObject(
	ctx context.Context,
	body io.Reader,
	contentType string,
	bucketName string,
	objectKey string,
) error {

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        body,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			return fmt.Errorf(
				"error while uploading object to %s; The object is too large: %w",
				bucketName, err,
			)

		}

		return fmt.Errorf(
			"couldn't upload object %s:%s: %w",
			bucketName, objectKey, err,
		)
	}

	if err = s.ObjectExists(ctx, bucketName, objectKey); err != nil {
		return fmt.Errorf(
			"failed attempt to wait for object %s:%s to exist: %w",
			bucketName, objectKey, err,
		)
	}

	return nil
}

// UploadFile uploads a file to bucket
func (s *service) UploadFile(ctx context.Context, bucketName, objectKey, filePath string) error {

	file, err := os.Open(filePath)
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
	err = s.PutObject(ctx, file, contentType, bucketName, objectKey)

	return err
}
