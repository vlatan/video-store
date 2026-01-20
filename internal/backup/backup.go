package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vlatan/video-store/internal/config"
	"github.com/vlatan/video-store/internal/integrations/r2"
)

type Service struct {
	config *config.Config
	r2s    r2.Service
}

// For your backup use case
var backupRoot string

func init() {
	var err error
	backupRoot, err = os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
}

// New creates a backup service
func New(cfg *config.Config, r2s r2.Service) *Service {
	return &Service{
		config: cfg,
		r2s:    r2s,
	}
}

// Run dumps a database to file and uploads that file to a bucket
func (s *Service) Run(ctx context.Context) error {

	// Print empty line at the end
	defer fmt.Println()

	log.Println("Backup service running...")

	dbDump := fmt.Sprintf("backup-%v.bak", time.Now().Format("2006-01-02T15-04"))
	if err := s.DumpDatabase(dbDump); err != nil {
		return err
	}

	log.Println("Database dumped.")

	archive := fmt.Sprintf("%s.tar.gz", dbDump)
	if err := s.ArchiveFiles(archive, dbDump); err != nil {
		return err
	}

	log.Println("Database compressed.")

	if err := s.r2s.UploadFile(
		ctx,
		s.config.R2BackupBucketName,
		backupRoot,
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

	// The dump command
	cmd := exec.Command("pg_dump",
		"-h", s.config.DBHost,
		"-p", strconv.Itoa(s.config.DBPort),
		"-U", s.config.DBUsername,
		"-d", s.config.DBDatabase,
		"-Fc", // compressed
		"-f", dest,
	) // #nosec G204

	// Set password via environment variable
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.config.DBPassword))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf(
			"pg_dump failed: %v\nstderr: %s\nstdout: %s",
			err, stderr.String(), stdout.String(),
		)
	}

	return nil
}

// ArchiveFiles tars files into an archive and gzips the archive
func (s *Service) ArchiveFiles(dest string, files ...string) error {

	// Create destination file
	destFile, err := r2.SecureCreate(backupRoot, dest)
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
		srcFile, err := r2.SecureOpen(backupRoot, src)
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
