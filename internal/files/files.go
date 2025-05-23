package files

import (
	"crypto/md5"
	"factual-docs/web"
	"fmt"
	"io/fs"
	"log"
	"regexp"
	"strings"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)

type fileInfo struct {
	bytes     []byte
	mediatype string
	Etag      string
}

type StaticFiles map[string]fileInfo

func NewStaticFiles() StaticFiles {
	m := minify.New()
	validJS := regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$")

	m.AddFunc("text/css", css.Minify)
	m.AddFuncRegexp(validJS, js.Minify)

	sf := make(StaticFiles)
	sf.ParseStaticFiles(m, "static")

	return sf
}

// Create minified versions of the static files and cache them in memory.
func (sf StaticFiles) ParseStaticFiles(m *minify.M, dir string) {
	// function used to process each file/dir in the root, including the root
	walkDirFunc := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// skip directories
		if info.IsDir() {
			return nil
		}

		// skip minified files
		if strings.Contains(info.Name(), ".min.") {
			return nil
		}

		// read the file
		b, err := fs.ReadFile(web.Files, path)
		if err != nil {
			return err
		}

		// split the file path on dot
		pathParts := strings.Split(path, ".")

		// set media type
		var mediatype string
		switch pathParts[1] {
		case "css":
			mediatype = "text/css"
		case "js":
			mediatype = "application/javascript"
		default:
			return nil
		}

		// minify the content
		b, err = m.Bytes(mediatype, b)
		if err != nil {
			return err
		}

		// create Etag as a hexadecimal md5 hash of the file content
		etag := fmt.Sprintf("%x", md5.Sum(b))

		// save all the data in the struct
		sf[info.Name()] = fileInfo{b, mediatype, etag}

		return nil
	}

	// walk the directory and process each file
	if err := fs.WalkDir(web.Files, dir, walkDirFunc); err != nil {
		log.Println(err)
	}
}
