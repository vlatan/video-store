package r2

import (
	"context"
	"errors"
	"factual-docs/internal/config"
	"fmt"
	"io"
	"log"
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
	// UploadFile uploads a file to bucket
	UploadFile(ctx context.Context, bucketName, objectKey, fileName string) error
	PutObject(
		ctx context.Context,
		content io.Reader,
		contentType, bucketName, objectKey string,
	) error
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

// PutObject puts object to bucket having the content ready
func (s *service) PutObject(
	ctx context.Context,
	content io.Reader,
	contentType, bucketName, objectKey string,
) error {

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey),
		Body:        content,
		ContentType: aws.String(contentType),
	})

	return err
}

// UploadFile uploads a file to bucket
func (s *service) UploadFile(ctx context.Context, bucketName, objectKey, fileName string) error {

	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("couldn't open file %s to upload: %w", fileName, err)
	}
	defer file.Close()

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
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
			"couldn't upload file %s to %s:%s: %v",
			fileName, bucketName, objectKey, err,
		)
	}

	err = s3.NewObjectExistsWaiter(s.client).Wait(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		},
		time.Minute,
	)

	if err != nil {
		return fmt.Errorf(
			"failed attempt to wait for object %s to exist: %w",
			objectKey, err,
		)
	}

	return nil
}
