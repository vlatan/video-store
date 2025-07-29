package backup

import (
	"bytes"
	"context"
	"errors"
	"factual-docs/internal/shared/config"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type Service struct {
	config   *config.Config
	s3Client *s3.Client
}

// New creates a backup service
func New(ctx context.Context) *Service {

	sdkConfig, err := awsConfig.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
	}

	return &Service{
		config:   config.New(),
		s3Client: s3.NewFromConfig(sdkConfig),
	}
}

// Run dumps a database to file and uploads that file to S3
func (s *Service) Run(ctx context.Context) error {

	fileName := fmt.Sprintf("backup-%v", time.Now().Format("2006-01-02-15:04"))
	err := s.DumpDatabase(fileName)
	if err != nil {
		return err
	}

	err = s.UploadFile(ctx, s.config.AwsBucketName, fileName, fileName)
	if err != nil {
		return err
	}

	return nil
}

// DumpDatabase dumps a database to file
func (s *Service) DumpDatabase(outputPath string) error {

	// Database URL
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		s.config.DBUsername,
		s.config.DBPassword,
		s.config.DBHost,
		s.config.DBPort,
		s.config.DBDatabase,
	)

	cmd := exec.Command("pg_dump", dbUrl, "-f", outputPath)

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("pg_dump failed: %v\nstderr: %s\nstdout: %s",
			err, stderr.String(), stdout.String())
	}

	return nil
}

// Upload a file to S3 bucket
func (s *Service) UploadFile(ctx context.Context, bucketName, objectKey, fileName string) error {

	// Open the file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("couldn't open file %s to upload: %v", fileName, err)
	}
	defer file.Close()

	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})

	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			return fmt.Errorf(
				"error while uploading object to %s; The object is too large: %v",
				bucketName, err,
			)

		}

		return fmt.Errorf(
			"couldn't upload file %s to %s:%s: %v",
			fileName, bucketName, objectKey, err,
		)
	}

	err = s3.NewObjectExistsWaiter(s.s3Client).Wait(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
		},
		time.Minute,
	)

	if err != nil {
		return fmt.Errorf("failed attempt to wait for object %s to exist", objectKey)
	}

	return nil
}
