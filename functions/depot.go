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

// NewFuncDepot creates a new FuncDepot under the given `namespace`
func NewFuncDepot(namespace string) *FuncDepot {
  return &FuncDepot { namespace, make(map[string]reflect.Value) }
}

// Get returns the function associated with the given key. The function
// is wrapped as reflect.Value so reflection can be used to determine
// attributes about this function
func (fc *FuncDepot) Get(key string) (reflect.Value, bool) {
  f, ok := fc.depot[key]
  return f, ok
}

// Set stores the function under the name `key`
func (fc *FuncDepot) Set(key string, v interface {}) {
  fc.depot[key] = reflect.ValueOf(v)
}

