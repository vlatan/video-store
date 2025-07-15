package static

import (
	"factual-docs/internal/models"
	"factual-docs/internal/shared/config"
	"regexp"
	"sync"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

type Service struct {
	sf     models.StaticFiles
	config *config.Config
}

var validJS = regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$")

var (
	sfInstance *Service
	once       sync.Once
)

// Create new files service
func New(config *config.Config) *Service {
	once.Do(func() {
		m := minify.New()
		m.AddFunc("text/css", css.Minify)
		m.AddFuncRegexp(validJS, js.Minify)

		sfInstance = &Service{
			sf:     parseStaticFiles(m, "static"),
			config: config,
		}
	})

	return sfInstance
}

// Returns the static files map
func (s *Service) GetStaticFiles() models.StaticFiles {
	return s.sf
}
