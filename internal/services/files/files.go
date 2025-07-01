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
	MediaType string
	Etag      string
}

type StaticFiles map[string]FileInfo

var validJS = regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$")

func New() StaticFiles {
	m := minify.New()

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
		var mediaType string
		switch fileType {
		case "css":
			mediaType = "text/css"
		case "js":
			mediaType = "application/javascript"
		}

		// Minify the content (only CSS or JS)
		if mediaType != "" {
			b, err = m.Bytes(mediaType, b)
			if err != nil {
				return err
			}
		}

		// Create Etag as a hexadecimal md5 hash of the file content
		etag := fmt.Sprintf("%x", md5.Sum(b))

		// Store empty bytes array if this is not CSS or JS
		if mediaType == "" {
			b = make([]byte, 0)
		}

		// Ensure the name starts with "/"
		name := path
		if !strings.HasPrefix(name, "/") {
			name = "/" + name
		}

		// Save all the data in the struct
		sf[name] = FileInfo{b, mediaType, etag}

		return nil
	}

	// Walk the directory and process each file
	if err := fs.WalkDir(web.Files, dir, walkDirFunc); err != nil {
		log.Println(err)
	}
}
