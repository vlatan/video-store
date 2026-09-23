package users

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestGetPageNum(t *testing.T) {
	tests := []struct {
		name, page string
		expected   int
	}{
		{"empty page", "", 1},
		{"valid page", "5", 5},
		{"invlaid page", "foo", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", fmt.Sprintf("/?page=%s", tt.page), nil)
			if got := GetPageNum(req); got != tt.expected {
				t.Errorf("got %d, want %d", got, tt.expected)
			}
		})
	}
}
