package hash

import (
  "github.com/lestrrat/go-xslate/functions"
)

var depot = functions.NewFuncDepot("hash")
func init () {
  depot.Set("Keys", Keys)
}

func Keys(m map[interface{}]interface{}) []interface{} {
  l := make([]interface {}, len(m))
  i := 0
  for k, _ := range m {
    l[i] = k
    i++
  }
  return l
}

func Depot() *functions.FuncDepot {
  return depot
}

