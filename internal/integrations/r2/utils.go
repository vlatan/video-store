package r2

import (
	"errors"
	"os"
)

type SecureFile struct {
	*os.File
	root *os.Root
}

func (sf *SecureFile) Close() error {
	fileErr := sf.File.Close()
	rootErr := sf.root.Close()
	return errors.Join(fileErr, rootErr)
}

// SecureOpen opens a file with a given root
func SecureOpen(rootPath, filename string) (*SecureFile, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	file, err := root.Open(filename)
	if err != nil {
		rootErr := root.Close() // Close the root if file open fails
		return nil, errors.Join(err, rootErr)
	}

	return &SecureFile{file, root}, nil
}

func SecureCreate(rootPath, filename string) (*SecureFile, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	file, err := root.Create(filename)
	if err != nil {
		rootErr := root.Close() // Close the root if file open fails
		return nil, errors.Join(err, rootErr)
	}

	return &SecureFile{file, root}, nil
}
