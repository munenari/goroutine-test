package main

import (
	"errors"
	"sync"
	"time"
)

func main() {}

type (
	worker struct {
		pool *sync.Pool
		job  func([]byte) error
	}
)

func (x worker) Do() error {
	b := x.pool.Get()
	defer x.pool.Put(b)
	bb, ok := b.([]byte)
	if !ok {
		return errors.New("pool returned invalid buffer")
	}
	return x.job(bb)
}

var (
	smallPool = &sync.Pool{
		New: func() any {
			return make([]byte, 1024) // 1KB
		},
	}
	bigPool = &sync.Pool{
		New: func() any {
			return make([]byte, 1024*1024) // 1MB
		},
	}
)

func lightJob(b []byte) error {
	_ = b
	time.Sleep(10 * time.Millisecond) // 10ms
	return nil
}

func heavyJob(b []byte) error {
	_ = b
	time.Sleep(100 * time.Millisecond) // 100ms
	return nil
}
