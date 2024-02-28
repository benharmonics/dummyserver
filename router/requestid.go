package router

import "sync/atomic"

type requestIDGenerator struct{ n atomic.Uint64 }

func (gen *requestIDGenerator) next() uint64 {
	ret := gen.n.Load()
	gen.n.Add(1)
	return ret
}
