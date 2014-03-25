package array

import (
  "github.com/lestrrat/go-xslate/functions"
)

var depot = functions.NewFuncDepot("array")
func init () {
  depot.Set("Item", Item)
  depot.Set("Size", Size)
}

// Item returns the `i`-th item in the list
func Item(l []interface{}, i int) interface{} {
  return l[i]
}

// Size returns the size of the list
func Size(l []interface {}) int {
  return len(l)
}

// Depot returns the FuncDepot for "array"
func Depot() *functions.FuncDepot {
  return depot
}

