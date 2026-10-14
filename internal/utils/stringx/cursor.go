package stringx

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// DecodeCursor decodes base64 string, splits the string on comma
// and returns a slice of strings.
func DecodeCursor(cursor string) ([]string, error) {
	decodedCursor, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, fmt.Errorf("could not decode the cursor; %w", err)
	}
	return strings.Split(string(decodedCursor), ","), nil
}
