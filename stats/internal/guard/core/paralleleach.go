package core

import (
	"runtime"
	"sync"
)

// ParallelEach runs fn(0..n-1) on at most GOMAXPROCS (the CPU count unless
// set lower) goroutines and returns once every call has.
func ParallelEach(n int, fn func(int)) {
	next := make(chan int)
	var wg sync.WaitGroup
	for range min(n, runtime.GOMAXPROCS(0)) {
		wg.Go(func() {
			for i := range next {
				fn(i)
			}
		})
	}
	for i := range n {
		next <- i
	}
	close(next)
	wg.Wait()
}
