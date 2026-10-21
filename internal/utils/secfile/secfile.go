package secfile

import (
	"errors"
	"os"
)

type secureFile struct {
	*os.File
	root *os.Root
}

func (f *secureFile) Close() error {
	return errors.Join(f.File.Close(), f.root.Close())
}

// SecureOpen opens a file with a given root
func Open(rootPath, filename string) (*secureFile, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	file, err := root.Open(filename)
	if err != nil {
		rootErr := root.Close() // Close the root if file open fails
		return nil, errors.Join(err, rootErr)
	}

	return &secureFile{file, root}, nil
}

func Create(rootPath, filename string) (*secureFile, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	file, err := root.Create(filename)
	if err != nil {
		rootErr := root.Close() // Close the root if file open fails
		return nil, errors.Join(err, rootErr)
	}

	return &secureFile{file, root}, nil
}
