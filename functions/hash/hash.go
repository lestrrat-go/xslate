package hash

import (
	"github.com/lestrrat-go/xslate/functions"
)

var depot = functions.NewFuncDepot("hash")

func init() {
	depot.Set("Keys", Keys)
}

// Keys returns the list of keys in this map. You can use this from a template
// like so `[% FOREACH key IN hash.Keys(mymap) %]...[% END %]` or
func Keys(m map[interface{}]interface{}) []interface{} {
	l := make([]interface{}, len(m))
	i := 0
	for k := range m {
		l[i] = k
		i++
	}
	return l
}

// Depot returns the Depot for hash package
func Depot() *functions.FuncDepot {
	return depot
}
