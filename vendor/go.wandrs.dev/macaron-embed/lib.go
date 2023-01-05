package embed

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"

	"gopkg.in/macaron.v1"
)

type EmbeddedFileSystem struct {
	FS fs.FS
}

var _ macaron.TemplateFileSystem = &EmbeddedFileSystem{}

func (e EmbeddedFileSystem) ListFiles() []macaron.TemplateFile {
	var files []macaron.TemplateFile
	_ = fs.WalkDir(e.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		data, err := fs.ReadFile(e.FS, path)
		if err != nil {
			return err
		}
		ext := filepath.Ext(d.Name())
		key := d.Name()
		name := key[0 : len(key)-len(ext)]
		files = append(files, macaron.NewTplFile(name, data, ext))
		return nil
	})
	return files
}

func (e EmbeddedFileSystem) Get(s string) (io.Reader, error) {
	var filename string
	err := fs.WalkDir(e.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		ext := filepath.Ext(d.Name())
		key := d.Name()
		if key[0:len(key)-len(ext)] == s {
			filename = path
			return errors.New("found")
		}
		return nil
	})
	if filename != "" {
		return e.FS.Open(filename)
	}
	return nil, err
}
