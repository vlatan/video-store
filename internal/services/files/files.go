package files

import (
	"factual-docs/internal/services/config"
	"regexp"
	"sync"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

type FileInfo struct {
	Bytes     []byte
	MediaType string
	Etag      string
}

type StaticFiles map[string]FileInfo

type Service struct {
	sf     StaticFiles
	config *config.Config
}

var validJS = regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$")

var (
	sfInstance *Service
	once       sync.Once
)

func New(config *config.Config) *Service {
	once.Do(func() {
		m := minify.New()
		m.AddFunc("text/css", css.Minify)
		m.AddFuncRegexp(validJS, js.Minify)

		sf := ParseStaticFiles(m, "static")

		sfInstance = &Service{
			sf:     sf,
			config: config,
		}
	})

	return sfInstance
}

func (s *Service) GetStaticFiles() StaticFiles {
	return s.sf
}
