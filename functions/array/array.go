package array

import (
  "github.com/lestrrat/go-xslate/functions"
)

var depot = functions.NewFuncDepot("array")
func init () {
  depot.Set("Item", Item)
}

func Item(l []interface{}, i int) interface{} {
  return l[i]
}

func Depot() *functions.FuncDepot {
  return depot
}

