package time

import (
  "time"
  "github.com/lestrrat/go-xslate/functions"
)

var depot = functions.NewFuncDepot("time")
func init () {
  depot.Set("After", time.After)
  depot.Set("Sleep", time.Sleep)
  depot.Set("Since", time.Since)
  depot.Set("Now", time.Now)
  depot.Set("ParseDuration", time.ParseDuration)
  depot.Set("Since", time.Since)
}

func Depot() *functions.FuncDepot {
  return depot
}