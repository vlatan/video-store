package static

import (
	"time"
)

type fileInfo struct {
	Bytes      []byte
	Compressed []byte
	MediaType  string
	ModTime    time.Time
	Etag       string
}

type staticFiles map[string]*fileInfo

type Service struct {
	staticFiles staticFiles
}
