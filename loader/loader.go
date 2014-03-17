/*
The loader packages abstracts the top-level xslate package from the job of
loading the bytecode from a key value.
*/
package loader

import (
  "errors"

  "github.com/lestrrat/go-xslate/vm"
)

// Teh ByteCodeLoader loads the ByteCode given a key
type ByteCodeLoader interface {
  Load(string) (*vm.ByteCode, error)
}

// The template loader loads the template string given a key
type TemplateLoader interface {
  Load(string) ([]byte, error)
}

var ErrTemplateNotFound = errors.New("Specified template was not found")

