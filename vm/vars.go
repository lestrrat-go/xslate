package vm

import (
  "fmt"
)

// Vars represents the variables passed into the Virtual Machine
type Vars map[string]interface {}

func NewVarsMerged(left Vars, right Vars) Vars {
  if left == nil && right != nil {
    return right
  } else if left != nil && right == nil {
    return left
  } else if left == nil && right == nil {
    return Vars {}
  }

  ret := Vars {}
  for k, v := range left {
    ret[k] = v
  }
  for k, v := range right {
    ret[k] = v
  }
  return ret
}

// Set sets the variable stored in slot `x`
func (v Vars) Set(k string, x interface {}) {
  v[k] = x
}

// Get returns the variable stored in slot `x`
func (v Vars) Get(k interface {}) (interface{}, bool) {
  key := fmt.Sprintf("%s", k)
  x, ok := v[key]
  return x, ok
}


