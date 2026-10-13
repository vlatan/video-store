package static

import "log/slog"

// AddVersion adds version query string to file
func (s *Service) AddVersion(path string) string {
	if fi, ok := s.staticFiles[path]; ok && fi.Etag != "" {
		return path + "?v=" + fi.Etag
	}
	slog.Warn("file not versioned", "path", path)
	return path
}
