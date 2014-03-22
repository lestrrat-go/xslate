package loader

import (
  "errors"
  "io/ioutil"
  "os"
  "path/filepath"
  "time"
)

// ErrAbsolutePathNotAllowed is returned when the given path is not a
// relative path. As of this writing, Xslate does not allow you to load 
// templates by absolute path, but this probably should be configurable
var ErrAbsolutePathNotAllowed = errors.New("error: Absolute paths are not allowed")

// FileTemplateFetcher is a TemplateFetcher that loads template strings
// in the file system.
type FileTemplateFetcher struct {
  Paths []string
}

// NewFileTemplateFetcher creates a new struct. `paths` must give us the
// directories for us to look the templates in
func NewFileTemplateFetcher(paths []string) (*FileTemplateFetcher, error) {
  l := &FileTemplateFetcher {
    Paths: make([]string, len(paths)),
  }
  for k, v := range paths {
    abs, err := filepath.Abs(v)
    if err != nil {
      return nil, err
    }
    l.Paths[k] = abs
  }
  return l, nil
}

// FetchTemplate returns a TemplateSource representing the template at path
// `path`. Paths are serached relative to the paths given to NewFileTemplateFetcher()
func (l *FileTemplateFetcher) FetchTemplate(path string) (TemplateSource, error) {
  if filepath.IsAbs(path) {
    return nil, ErrAbsolutePathNotAllowed
  }

  for _, dir := range l.Paths {
    fullpath := filepath.Join(dir,  path)

    _, err := os.Stat(fullpath)
    if err != nil {
      continue
    }

    return NewFileSource(fullpath), nil
  }
  return nil, ErrTemplateNotFound
}

// FileSource is a TemplateSource variant that holds template information
// in a file.
type FileSource struct {
  Path string
}

// NewFileSource creates a new FileSource
func NewFileSource(path string) *FileSource {
  return &FileSource { path }
}

// LastModified returns time when the target template file was last modified
func (s *FileSource) LastModified() (time.Time, error) {
  fi, err := os.Stat(s.Path)
  if err != nil {
    return time.Time {}, err
  }

  return fi.ModTime(), nil
}

// Bytes returns the bytes in teh template file
func (s *FileSource) Bytes() ([]byte, error) {
  fh, err := os.Open(s.Path)
  if err != nil {
    return nil, err
  }
  return ioutil.ReadAll(fh)
}
