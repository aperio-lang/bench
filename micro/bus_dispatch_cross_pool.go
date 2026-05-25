// Go equivalent of bus_dispatch_cross_pool.hl.
//
// Two goroutines: main publishes, a worker subscribes. Cross-
// thread delivery via a buffered channel (cap 64, mirroring
// Hale's bounded io pool queue depth — unbuffered would pin
// publish to drain rate per message; large buffer would let
// the publisher dump and run, defeating the purpose).
//
// runtime.LockOSThread on the subscriber ensures the worker
// goroutine stays on a distinct OS thread for the duration of
// the bench — closest Go analogue of "cooperative(pool = io)
// has its own OS thread".
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Tick struct {
	N int
}

func main() {
	iters := 100000
	ch := make(chan Tick, 64)
	var wg sync.WaitGroup
	wg.Add(1)

	count := 0
	go func() {
		runtime.LockOSThread()
		defer wg.Done()
		for t := range ch {
			count++
			_ = t
		}
	}()

	t0 := time.Now()
	for i := 0; i < iters; i++ {
		ch <- Tick{N: i}
	}
	elapsed := time.Since(t0).Nanoseconds()
	close(ch)
	wg.Wait()

	fmt.Printf("iters=%d\n", iters)
	fmt.Printf("count=%d\n", count)
	fmt.Printf("elapsed_ns=%d\n", elapsed)
}
