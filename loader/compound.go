package loader

import (
  "github.com/lestrrat/go-xslate/vm"
)

type CompoundByteCodeLoader struct {
  Loaders []ByteCodeLoader
}

func (l *CompoundByteCodeLoader) Load(key string) (*vm.ByteCode, error) {
  for _, v := range l.Loaders {
    bc, err := v.Load(key)
    if err != nil {
      return bc, nil
    }
  }

  return nil, ErrTemplateNotFound
}

