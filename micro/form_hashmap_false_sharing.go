// Go equivalent of form_hashmap_false_sharing.hl.
//
// Two goroutines hammer disjoint keys (even / odd ids) of a
// shared map. Cross-goroutine access via sync.Mutex (the
// canonical Go shape for a shared mutating map; sync.Map is
// optimized for "many readers, occasional writer" which isn't
// this workload). runtime.LockOSThread on each goroutine pins
// them to distinct OS threads for the duration — closest
// analogue of Hale's cooperative pools each owning their own
// OS thread.
//
// Lock discipline matches Hale's `sync = serialized`: every
// mutate takes the per-map lock. Throughput is bounded by
// contention; this is the safe-but-slow baseline both
// languages share.
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Counter struct {
	Id    int
	Value int
}

type Registry struct {
	mu      sync.Mutex
	entries map[int]Counter
}

func (r *Registry) Set(c Counter) {
	r.mu.Lock()
	r.entries[c.Id] = c
	r.mu.Unlock()
}

func (r *Registry) Len() int {
	r.mu.Lock()
	n := len(r.entries)
	r.mu.Unlock()
	return n
}

func main() {
	perWriter := 100000
	total := perWriter * 2
	reg := &Registry{entries: make(map[int]Counter)}

	var wg sync.WaitGroup
	wg.Add(2)

	t0 := time.Now()

	// Goroutine A — even ids
	go func() {
		runtime.LockOSThread()
		defer wg.Done()
		for i := 0; i < perWriter; i++ {
			reg.Set(Counter{Id: i * 2, Value: i})
		}
	}()

	// Goroutine B — odd ids
	go func() {
		runtime.LockOSThread()
		defer wg.Done()
		for i := 0; i < perWriter; i++ {
			reg.Set(Counter{Id: i*2 + 1, Value: i})
		}
	}()

	wg.Wait()
	elapsed := time.Since(t0).Nanoseconds()

	fmt.Printf("per_writer=%d\n", perWriter)
	fmt.Printf("total=%d\n", reg.Len())
	fmt.Printf("elapsed_ns=%d\n", elapsed)
	_ = total
}
