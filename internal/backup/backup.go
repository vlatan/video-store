package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"factual-docs/internal/config"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

type Service struct {
	config *config.Config
	client *s3.Client
}

// New creates a backup service
func New(cfg *config.Config, client *s3.Client) *Service {
	return &Service{
		config: cfg,
		client: client,
	}
}

// Run dumps a database to file and uploads that file to a bucket
func (s *Service) Run(ctx context.Context) error {

	log.Println("Backup service running...")

	dbDump := fmt.Sprintf("backup-%v", time.Now().Format("2006-01-02T15-04"))
	if err := s.DumpDatabase(dbDump); err != nil {
		return err
	}

	log.Println("Database dumped.")

	archive := fmt.Sprintf("%s.tar.gz", dbDump)
	if err := s.ArchiveFiles([]string{dbDump}, archive); err != nil {
		return err
	}

	log.Println("Database compressed.")

	if err := s.UploadFile(
		ctx,
		s.config.R2BackupBucketName,
		archive,
		archive,
	); err != nil {
		return err
	}

	log.Println("Database uploaded to bucket.")
	log.Println("Backup finished successfully.")

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
func (s *Service) ArchiveFiles(files []string, dest string) error {

	// Create destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dest, err)
	}
	defer destFile.Close()

	// Create gzip writer with maximum compression
	gzipWriter, err := gzip.NewWriterLevel(destFile, gzip.BestCompression)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %w", err)
	}
	defer gzipWriter.Close()

	// Create tar writer (chained with gzip writer)
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Go over the files
	for _, src := range files {

		// Open the source file
		srcFile, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("failed to open source file %s: %w", src, err)
		}
		defer srcFile.Close()

		// Get file info for the tar header
		srcInfo, err := srcFile.Stat()
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", src, err)
		}

		// Create tar header
		header := &tar.Header{
			Name:    filepath.Base(src),
			Size:    srcInfo.Size(),
			Mode:    int64(srcInfo.Mode()),
			ModTime: srcInfo.ModTime(),
		}

		// Write header to tar
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for file %s: %w", src, err)
		}

		// Copy file content
		if _, err := io.Copy(tarWriter, srcFile); err != nil {
			return fmt.Errorf("failed to copy file content for file %s: %w", src, err)
		}
	}

	return nil
}

// UploadFile uploads a file to bucket
func (s *Service) UploadFile(ctx context.Context, bucketName, objectKey, fileName string) error {

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
