package gemini

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// The dir in which the data will be fetched
const dataDir = "/tmp/data"
const framesDir = "/tmp/data/frames"
const outputStem = "/tmp/data/output"

// extractAudio extracts the audio from a given YT video.
func extractAudio(videoID string) error {

	cmdArgs := []string{
		"--keep-video",
		"--extract-audio",
		"--audio-format",
		"mp3",
		videoID,
		"-o",
		outputStem + ".mp3",
	}

	var errBuf bytes.Buffer
	cmd := exec.Command("yt-dlp", cmdArgs...)
	cmd.Stderr = &errBuf

	// #nosec G204
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("STDERR: %s\n%w", errBuf.String(), err)
	}

	return nil
}

// extractImages extracts frames from a video
func extractImages(videoFilePath, outputDir string) error {

	// Get 1 image per second, the first 210 seconds
	cmdArgs := []string{
		"-threads",
		"1",
		"-loglevel",
		"error",
		"-i",
		videoFilePath,
		"-t",
		"210",
		"-vf",
		"fps=1",
		filepath.Join(outputDir, "first_%04d.png"),
	}

	var errBuf bytes.Buffer
	cmd := exec.Command("ffmpeg", cmdArgs...)
	cmd.Stderr = &errBuf

	// #nosec G204
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("STDERR: %s\n%w", errBuf.String(), err)
	}

	// Get video duration
	cmdArgs = []string{
		"-v",
		"error",
		"-show_entries",
		"format=duration",
		"-of",
		"default=noprint_wrappers=1:nokey=1",
		videoFilePath,
	}

	errBuf.Reset()
	cmd = exec.Command("ffprobe", cmdArgs...)
	cmd.Stderr = &errBuf

	// #nosec G204
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("STDERR: %s\n%w", errBuf.String(), err)
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return err
	}

	duration := int(math.Round(durationFloat))

	// Get 1 image per second, the last 210 seconds
	cmdArgs = []string{
		"-threads",
		"1",
		"-loglevel",
		"error",
		"-ss",
		fmt.Sprintf("%d", duration-210),
		"-i",
		videoFilePath,
		"-vf",
		"fps=1",
		filepath.Join(outputDir, "last_%04d.png"),
	}

	errBuf.Reset()
	cmd = exec.Command("ffmpeg", cmdArgs...)
	cmd.Stderr = &errBuf

	// #nosec G204
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("STDERR: %s\n%w", errBuf.String(), err)
	}

	return nil
}

// findVideo finds a video on disk with a given stem
func findVideo() (string, error) {

	exts := []string{".mp4", ".mkv", ".webm", ".avi", ".mov"}
	for _, ext := range exts {
		path := outputStem + ext
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no video found for stem %q", outputStem)
}

// extractMedia extracts media given a YT video ID,
// namely audio and images.
func extractMedia(videoID string) error {

	// Create dirs
	if err := os.MkdirAll(framesDir, 0750); err != nil {
		return fmt.Errorf("can't make directories; %w", err)
	}

	if err := extractAudio(videoID); err != nil {
		return fmt.Errorf("can't extract audio; %w", err)
	}

	videoFilePath, err := findVideo()
	if err != nil {
		return fmt.Errorf("can't find extracted video; %w", err)
	}

	if err := extractImages(videoFilePath, framesDir); err != nil {
		return fmt.Errorf("can't extract images; %w", err)
	}

	return nil
}
