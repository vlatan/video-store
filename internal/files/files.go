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

type FileInfo struct {
	Bytes     []byte
	Mediatype string
	Etag      string
}

type StaticFiles map[string]FileInfo

func New() StaticFiles {
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
	// Function used to process each file/dir in the root, including the root
	walkDirFunc := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip minified files
		if strings.Contains(info.Name(), ".min.") {
			return nil
		}

		// Read the file
		b, err := fs.ReadFile(web.Files, path)
		if err != nil {
			return err
		}

		// Get the fyle type
		fileType := strings.Split(info.Name(), ".")[1]

		// Set media type
		var mediatype string
		switch fileType {
		case "css":
			mediatype = "text/css"
		case "js":
			mediatype = "application/javascript"
		default:
			return nil
		}

		// Minify the content
		b, err = m.Bytes(mediatype, b)
		if err != nil {
			return err
		}

		// Create Etag as a hexadecimal md5 hash of the file content
		etag := fmt.Sprintf("%x", md5.Sum(b))

		// Save all the data in the struct
		sf[path] = FileInfo{b, mediatype, etag}

		return nil
	}

	// Walk the directory and process each file
	if err := fs.WalkDir(web.Files, dir, walkDirFunc); err != nil {
		log.Println(err)
	}
}
