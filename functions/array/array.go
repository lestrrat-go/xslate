package array

import (
  "github.com/lestrrat/go-xslate/functions"
)

var depot = functions.NewFuncDepot("array")
func init () {
  depot.Set("Item", Item)
}

// Item returns the `i`-th item in the list
func Item(l []interface{}, i int) interface{} {
  return l[i]
}

// Depot returns the FuncDepot for "array"
func Depot() *functions.FuncDepot {
  return depot
}

