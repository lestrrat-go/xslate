package vm

import (
  "fmt"
)

type Vars map[string]interface {}

func (v Vars) Set(k string, x interface {}) {
  v[k] = x
}

func (v Vars) Get(k interface {}) (interface{}, bool) {
  key := fmt.Sprintf("%s", k)
  x, ok := v[key]
  return x, ok
}


