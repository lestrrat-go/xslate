package vm

import (
	"fmt"
)

// Vars represents the variables passed into the Virtual Machine
type Vars map[string]interface{}

// Set sets the variable stored in slot `x`
func (v Vars) Set(k string, x interface{}) {
	v[k] = x
}

// Get returns the variable stored in slot `x`
func (v Vars) Get(k interface{}) (interface{}, bool) {
	key := fmt.Sprintf("%s", k)
	x, ok := v[key]
	return x, ok
}
