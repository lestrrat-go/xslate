package loader

import (
  "errors"
  "io/ioutil"
  "os"
  "path/filepath"
)

var ErrAbsolutePathNotAllowed = errors.New("Absolute paths are not allowed")
type LoadFile struct {
  Paths []string
}

func NewLoadFile(paths []string) (*LoadFile, error) {
  l := &LoadFile {
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

func (l *LoadFile) Load(path string) ([]byte, error) {
  if filepath.IsAbs(path) {
    return nil, ErrAbsolutePathNotAllowed
  }

  for _, dir := range l.Paths {
    fullpath := filepath.Join(dir,  path)
    fh, err := os.Open(fullpath)
    if err != nil {
      continue
    }

    return ioutil.ReadAll(fh)
  }

  return nil, ErrTemplateNotFound
}