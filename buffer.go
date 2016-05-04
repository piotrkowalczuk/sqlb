package sqlb

import (
	"bytes"
	"sync"
)

type buffer struct {
	pool sync.Pool
}

func newBuffer() *buffer {
	return &buffer{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(nil)
			},
		},
	}
}

func (b *buffer) get() *bytes.Buffer {
	return b.pool.Get().(*bytes.Buffer)
}

func (b *buffer) put(buf *bytes.Buffer) {
	buf.Reset()
	b.pool.Put(buf)
}
