package rvpool

import (
	"sync"
)

// render vars pool
var pool = sync.Pool{
	New: allocRenderVars,
}

func allocRenderVars() interface{} {
	return map[string]interface{}{}
}

func Get() map[string]interface{} {
	return pool.Get().(map[string]interface{})
}

func Release(m map[string]interface{}) {
	pool.Put(m)
}
