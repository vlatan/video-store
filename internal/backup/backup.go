package backup

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"factual-docs/internal/shared/config"
	"fmt"
	"io"
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

	dbDump := fmt.Sprintf("backup-%v", time.Now().Format("2006-01-02T15-04"))
	if err := s.DumpDatabase(dbDump); err != nil {
		return err
	}

	cDump := fmt.Sprintf("%s.gz", dbDump)
	if err := s.CompressFile(dbDump, cDump); err != nil {
		return err
	}

	if err := s.UploadFile(ctx, s.config.AwsBucketName, cDump, cDump); err != nil {
		return err
	}

	return nil
}

// DumpDatabase dumps a database to file
func (s *Service) DumpDatabase(dest string) error {

	// Database URL
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		s.config.DBUsername,
		s.config.DBPassword,
		s.config.DBHost,
		s.config.DBPort,
		s.config.DBDatabase,
	)

	cmd := exec.Command("pg_dump", dbUrl, "-f", dest)

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

// Compress compresses a file
func (s *Service) CompressFile(src, dest string) error {

	// Open the original file for reading
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open the file: %w", err)
	}
	defer file.Close()

	// Create the destination gzip file
	gzipFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create gzip file: %w", err)
	}
	defer gzipFile.Close()

	// Create a gzip writer that writes to the destination file
	gzipWriter, err := gzip.NewWriterLevel(gzipFile, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	// Copy the content from the source file to the gzip writer
	_, err = io.Copy(gzipWriter, file)
	if err != nil {
		return fmt.Errorf("failed to copy data to gzip writer: %w", err)
	}

	return nil
}

// UploadFile uploads a file to S3 bucket
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
		return fmt.Errorf(
			"failed attempt to wait for object %s to exist: %v",
			objectKey, err,
		)
	}

	return nil
}
