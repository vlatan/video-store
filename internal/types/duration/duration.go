package duration

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Custom string type used to convert string duration to desirable format
type ISO8601 string

// Valid ISO time format
var validISO8601 = regexp.MustCompile(`(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?`)

// compile compiles ISO-8601 string to time components
func (d ISO8601) compile() (map[string]int, error) {
	// Check for PT prefix
	if !strings.HasPrefix(string(d), "PT") {
		return nil, fmt.Errorf("invalid duration format: %s", d)
	}

	// Remove the PT prefix
	duration := strings.TrimPrefix(string(d), "PT")

	// Find the substrings (hours, minutes, seconds)
	matches := validISO8601.FindStringSubmatch(duration)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid duration format: %s", duration)
	}

	// Check for the matched regex groups
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	sec, _ := strconv.ParseFloat(matches[3], 64)
	seconds := int(sec)

	return map[string]int{
		"h": hours,
		"m": minutes,
		"s": seconds,
	}, nil
}

// Human converts ISO-8601 string to human readbale duration
func (d ISO8601) Human() string {
	t, err := d.compile()
	if err != nil {
		slog.Warn("failed to compile video duration", "duration", d)
		return ""
	}

	return fmt.Sprintf("%02d:%02d:%02d", t["h"], t["m"], t["s"])
}

// Parse converts ISO-8601 string to duration
func (d ISO8601) Parse() (time.Duration, error) {
	t, err := d.compile()
	if err != nil {
		return 0, err
	}

	duration := time.Duration(t["h"])*time.Hour +
		time.Duration(t["m"])*time.Minute +
		time.Duration(t["s"])*time.Second

	return duration, nil

}
