// internal/encoding/jsonx/pool.go
package jsonx

import (
	"bytes"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func GetBuffer() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func PutBuffer(b *bytes.Buffer) {
	if b.Cap() > 1<<20 { // evita pools gigantes (1MB)
		return
	}
	bufPool.Put(b)
}
