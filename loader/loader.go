package loader

import (
  "errors"
)

type Loader interface {
  Load(string) ([]byte, error)
}

var ErrTemplateNotFound = errors.New("Specified template was not found")