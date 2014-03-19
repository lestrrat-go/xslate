package loader

import (
  "errors"
  "io/ioutil"
  "os"
  "path/filepath"
  "time"
)

var ErrAbsolutePathNotAllowed = errors.New("Absolute paths are not allowed")

type FileTemplateFetcher struct {
  Paths []string
}

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

type FileSource struct {
  Path string
}

func NewFileSource(path string) *FileSource {
  return &FileSource { path }
}

func (s *FileSource) LastModified() (time.Time, error) {
  fi, err := os.Stat(s.Path)
  if err != nil {
    return time.Time {}, err
  }

  return fi.ModTime(), nil
}

func (s *FileSource) Bytes() ([]byte, error) {
  fh, err := os.Open(s.Path)
  if err != nil {
    return nil, err
  }
  return ioutil.ReadAll(fh)
}
