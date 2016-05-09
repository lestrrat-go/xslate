package vm

import (
	"fmt"
)

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

func (v *Vars) Reset() {
	for k := range *v {
		delete(*v, k)
	}
}
