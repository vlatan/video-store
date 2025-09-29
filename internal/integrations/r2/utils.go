package r2

import "os"

type SecureFile struct {
	*os.File
	root *os.Root
}

func (sf *SecureFile) Close() error {
	fileErr := sf.File.Close()
	rootErr := sf.root.Close()

	if fileErr != nil {
		return fileErr
	}
	return rootErr
}

// SecureOpen opens a file with a given root
func SecureOpen(rootPath, filename string) (*SecureFile, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}

	file, err := root.Open(filename)
	if err != nil {
		root.Close() // Clean up root if file open fails
		return nil, err
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
		root.Close() // Clean up root if file open fails
		return nil, err
	}

	return &SecureFile{file, root}, nil
}
