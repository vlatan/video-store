package gemini

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// The dir in which the data will be fetched
const dataDir = "data"
const framesDir = "data/frames"
const audioFile = "data/output.mp3"

// extractAudio extracts the audio from a given YT video.
func extractAudio(videoID string) error {

	cmdArgs := []string{
		"--keep-video",
		"--extract-audio",
		"--audio-format",
		"mp3",
		videoID,
		"-o",
		audioFile,
	}

	return exec.Command("yt-dlp", cmdArgs...).Run()
}

func extractImages(videoFilePath, outputDir string) error {

	// Get 1 image per second, the first 180 seconds
	cmdArgs := []string{
		"-loglevel",
		"error",
		"-i",
		videoFilePath,
		"-t",
		"180",
		"-vf",
		"fps=1",
		filepath.Join(outputDir, "first_%04d.png"),
	}

	if err := exec.Command("ffmpeg", cmdArgs...).Run(); err != nil {
		return err
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

	output, err := exec.Command("ffprobe", cmdArgs...).Output()
	if err != nil {
		return err
	}

	durationStr := strings.TrimSpace(string(output))
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return err
	}

	duration := int(math.Round(durationFloat))

	cmdArgs = []string{
		"-loglevel",
		"error",
		"-ss",
		fmt.Sprintf("%d", duration-60),
		"-i",
		videoFilePath,
		"-vf",
		"fps=1",
		filepath.Join(outputDir, "last_%04d.png"),
	}

	return exec.Command("ffmpeg", cmdArgs...).Run()
}

func findMergedVideo() (string, error) {

	stem := strings.TrimSuffix(audioFile, filepath.Ext(audioFile))
	videoExts := []string{".mp4", ".mkv", ".webm", ".avi", ".mov"}

	for _, ext := range videoExts {
		path := stem + ext
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no merged video found for stem %q", stem)
}

// extractMedia extracts media given a YT video ID,
// namely audio and images.
func extractMedia(videoID string) error {

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("can't get the current dir; %w", err)
	}

	root, err := os.OpenRoot(currentDir)
	if err != nil {
		return fmt.Errorf("can't open current dir as root; %w", err)
	}

	defer root.Close()

	// Create dirs
	if err := root.MkdirAll(framesDir, 0755); err != nil {
		return fmt.Errorf("can't make directories; %w", err)
	}

	if err := extractAudio(videoID); err != nil {
		return fmt.Errorf("can't extract audio; %w", err)
	}

	videoFilePath, err := findMergedVideo()
	if err != nil {
		return fmt.Errorf("can't find extracted video; %w", err)
	}

	if err := extractImages(videoFilePath, framesDir); err != nil {
		return fmt.Errorf("can't extract images; %w", err)
	}

	return nil
}
