// Go equivalent of form_hashmap_walk_large.hl.
//
// Pre-populate a map, then iterate via range — Go's built-in
// map iteration is the natural analogue of Hale's key_at /
// entry_at sweep. Sum the values to defeat dead-code
// elimination. Iteration order is hash-table order in both
// languages (Go intentionally randomizes seed per map; Hale
// is insertion-affected but deterministic per table state) —
// the comparison is throughput-per-entry, not order.
package main

import (
	"fmt"
	"time"
)

type Entry struct {
	Id  int
	Val int
}

func main() {
	n := 100000
	m := make(map[int]Entry, n)
	for i := 0; i < n; i++ {
		m[i] = Entry{Id: i, Val: i * 3}
	}

	t0 := time.Now()
	sum := 0
	for k, e := range m {
		sum += e.Val + (k - k)
	}
	elapsed := time.Since(t0).Nanoseconds()

	fmt.Printf("n=%d\n", n)
	fmt.Printf("len=%d\n", len(m))
	fmt.Printf("sum=%d\n", sum)
	fmt.Printf("elapsed_ns=%d\n", elapsed)
}
