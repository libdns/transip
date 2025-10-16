package client

import (
	"bytes"
	"sync"
)

type buf struct {
	bytes.Buffer
	closer func(x any)
}

func (b *buf) Close() error {
	b.Reset()
	b.closer(b)
	return nil
}

func NewBufPool() *sync.Pool {

	var pool = &sync.Pool{}

	pool.New = func() interface{} {
		return &buf{
			closer: pool.Put,
		}
	}

	return pool
}
