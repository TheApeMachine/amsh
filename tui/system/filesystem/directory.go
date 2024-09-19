package filesystem

import (
	"bytes"
	"io/fs"
	"os"
)

/*
Directory handles the directory structure of the filesystem.
*/
type Directory struct {
	path  string
	files []fs.DirEntry
	buf   *bytes.Buffer
}

/*
NewDirectory creates a new directory with the given path.
*/
func NewDirectory(path string) *Directory {
	return &Directory{
		path:  path,
		files: make([]fs.DirEntry, 0),
		buf:   bytes.NewBuffer([]byte{}),
	}
}

/*
Read the directory at path and write the contents to the buffer.
*/
func (directory *Directory) Read(p []byte) (n int, err error) {
	if directory.files, err = os.ReadDir(directory.path); err != nil {
		return 0, err
	}

	for _, file := range directory.files {
		directory.buf.WriteString(file.Name() + "\n")
	}

	return directory.buf.Read(p)
}
