package middlewares

import (
	"net/http"
	"strings"
)

// GetSrcIp returns the client IP
func remoteIP(r *http.Request) string {

	// Prioritize CF-Connecting-IP as recommended by Cloudflare
	srcIp := r.Header.Get("CF-Connecting-IP")

	// Fallback to True-Client-IP
	if srcIp == "" {
		srcIp = r.Header.Get("True-Client-IP")
	}

	// Fallback to X-Forwarded-For
	if srcIp == "" {
		xForwardedFor := r.Header.Get("X-Forwarded-For")
		parts := strings.Split(xForwardedFor, ",")
		srcIp = strings.TrimSpace(parts[0])
	}

	// Fallback to RemoteAddr
	if srcIp == "" {
		srcIp = r.RemoteAddr
	}

	return srcIp
}
