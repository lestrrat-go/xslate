package functions

import (
  "reflect"
)

// FuncDepot is a map of function name to it's real content
// wrapped in reflect.ValueOf()
type FuncDepot struct {
  namespace string
  depot     map[string]reflect.Value
}

func NewFuncDepot(namespace string) *FuncDepot {
  return &FuncDepot { namespace, make(map[string]reflect.Value) }
}

func (fc *FuncDepot) Get(key string) (reflect.Value, bool) {
  f, ok := fc.depot[key]
  return f, ok
}

func (fc *FuncDepot) Set(key string, v interface {}) {
  fc.depot[key] = reflect.ValueOf(v)
}

