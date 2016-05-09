package rbpool

import (
	"bytes"
	"sync"
)

// render buffer pool
var pool = sync.Pool{
	New: allocRenderBuffer,
}

func allocRenderBuffer() interface{} {
	return &bytes.Buffer{}
}

func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func Release(buf *bytes.Buffer) {
	buf.Reset()
	pool.Put(buf)
}
