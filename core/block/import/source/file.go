package source

import (
	"io"
	"os"
	"path/filepath"

	oserror "github.com/anyproto/anytype-heart/util/os"
)

type File struct {
	fileReaders map[string]io.ReadCloser
}

func NewFile() *File {
	return &File{}
}

func (f *File) GetFileReaders(importPath string, expectedExt []string, includeFiles []string) (map[string]io.ReadCloser, error) {
	shortPath := filepath.Clean(importPath)
	if !isFileAllowedToImport(shortPath, filepath.Ext(importPath), expectedExt, includeFiles) {
		return nil, nil
	}
	files := make(map[string]io.ReadCloser, 0)
	file, err := os.Open(importPath)
	if err != nil {
		return nil, oserror.TransformError(err)
	}
	files[shortPath] = file
	f.fileReaders = files
	return files, nil
}

func (f *File) Close() {
	for _, fileReader := range f.fileReaders {
		fileReader.Close()
	}
}
