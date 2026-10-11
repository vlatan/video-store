package static

import (
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
)

var validJS = regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$")

func New() (*Service, error) {

	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFuncRegexp(validJS, js.Minify)
	m.AddFunc("application/manifest+json", json.Minify)

	sf, err := loadStaticFiles(m, "static")
	if err != nil {
		return nil, err
	}

	return &Service{sf}, nil
}
