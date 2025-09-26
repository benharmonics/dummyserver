package router

import "sync"

type requestIDGenerator struct {
	current uint64
	mu      sync.Mutex
}

func (gen *requestIDGenerator) next() uint64 {
	gen.mu.Lock()
	defer gen.mu.Unlock()
	ret := gen.current
	gen.current++
	return ret
}
