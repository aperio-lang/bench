# Hale benchmark suite

Performance harness for Hale, comparing against Go / Node / Python
sibling implementations of the same workload shape. For the language
itself, see [hale-lang/hale](https://github.com/hale-lang/hale).
Baseline scaffold for Layer 3 (performance) per `spec/testing.md` in
the hale repo.
Microbenches isolate substrate primitives; one app-level bench
exercises a small mixed workload end-to-end. The runner is
shell + jq, no extra build dependencies.

Each `.hl` bench ships with sibling `.go`, `.js`, and `.py`
equivalents implementing the same shape as closely as possible.
The harness runs all of them, reports per-language medians plus
a `ratio_vs_hale` column, but **only Hale regressions** gate
exit code (per `spec/testing.md` Layer 3: "a regression in
hale-vs-X ratio is a developer signal, not a CI gate").

## Quickstart

```
# From repo root, with the CLI built (cargo build --release -p hale-cli):
./run.sh                       # run all + comparatives, exit 1 on Hale regression
./run.sh --iters=10            # more samples per bench (default 5)
./run.sh --bench=loop_overhead # one bench at a time
./run.sh --update-baselines    # overwrite baselines.json with new Hale medians
./run.sh --no-build            # skip rebuild step
./run.sh --no-comparative      # Hale only, skip go/node/python siblings
./run.sh --json                # emit only the JSON report to stdout
```

Each invocation writes a timestamped JSON report under
`results/` (gitignored).

## Cross-language comparative grid

Latest snapshot: **Hale v0.8.0** (2026-05-25), AMD Ryzen 7
9800X3D / x86_64 / Linux 6.18.

Read each cell as `<elapsed> (<ratio_vs_hale>×)` where
**ratio = sibling_time / hale_time**:

- ratio > 1 → sibling is **slower** than Hale by that factor.
- ratio < 1 → sibling is **faster** than Hale by 1/ratio.
- `—` → no sibling exists for this bench (Hale-only).

### Overhead microbenches

| Bench | Hale | Go | Node | Python |
|---|---:|---:|---:|---:|
| `loop_overhead`             | 12.07 ms | 19.46 ms (1.61×) | 20.84 ms (1.73×) | 3.42 s (283×) |
| `fn_call`                   | 16.64 ms | 7.69 ms (0.46×) | 3.20 ms (0.19×) | 383.36 ms (23.0×) |
| `locus_instantiation`       | 2.04 ms | 0.15 ms (0.08×) | 0.96 ms (0.47×) | 12.37 ms (6.07×) |
| `bus_dispatch`              | 9.17 ms | 0.05 ms (0.005×) | 0.30 ms (0.033×) | 1.10 ms (0.12×) |
| `bus_dispatch_heap_payload` | 5.08 ms | — | — | — |
| `bus_publish_shm_ring`      | 1.38 ms | — | — | — |
| `form_vec_push`             | 31.39 ms | 2.82 ms (0.09×) | 3.88 ms (0.12×) | 13.38 ms (0.43×) |
| `form_vec_get`              | 9.61 ms | 0.04 ms (0.004×) | 0.67 ms (0.07×) | 5.91 ms (0.61×) |
| `form_hashmap_set`          | 41.30 ms | 54.46 ms (1.32×) | 90.23 ms (2.18×) | 275.35 ms (6.67×) |
| `form_hashmap_get`          | 5.75 ms | 1.11 ms (0.19×) | 2.57 ms (0.45×) | 9.33 ms (1.62×) |

### Amortized microbenches

| Bench | Hale | Go | Node | Python |
|---|---:|---:|---:|---:|
| `vec_amortized`     | 1.04 ms | 1.27 ms (1.22×) | 3.06 ms (2.93×) | 13.58 ms (13.0×) |
| `fn_scratch_work`   | 0.42 ms | 0.45 ms (1.08×) | 0.97 ms (2.32×) | 2.59 ms (6.19×) |
| `coord_with_churn`  | 53.50 µs | 0.16 µs (0.003×) | 33.47 µs (0.63×) | 4.19 µs (0.08×) |

### Coordinated-workload microbenches

| Bench | Hale | Go | Node | Python |
|---|---:|---:|---:|---:|
| `tree_fanout`      | 24.61 µs | 8.02 µs (0.33×) | 257.84 µs (10.48×) | 579.64 µs (23.56×) |
| `pipeline_3stage`  | 7.14 ms | 0.22 ms (0.03×) | 1.15 ms (0.16×) | 6.93 ms (0.97×) |

### Cross-pool / cache microbenches (F.32)

| Bench | Hale | Go | Node | Python |
|---|---:|---:|---:|---:|
| `bus_dispatch_cross_pool`     | 11.85 ms | 6.08 ms (0.51×) | 40.82 ms (3.45×) | 82.33 ms (6.95×) |
| `form_hashmap_false_sharing`  | 11.95 ms | 10.12 ms (0.85×) | 85.83 ms (7.18×) | 36.69 ms (3.07×) |
| `form_hashmap_walk_large`     | 1.12 ms | 0.38 ms (0.34×) | 0.76 ms (0.68×) | 7.21 ms (6.43×) |

`form_hashmap_false_sharing` exercises the F.32-1γ-v1
`sync = lockfree` discipline (cell-level CAS, no rwlock,
no mutex; fixed cap). On this 2-core / cheap-payload
bench, γ-v1 closes the gap with Go's sync.Mutex map from
1.66× (under α serialized) to 1.18×. The annotation chain
shows the F.32-1 alternatives in order of measured perf
on this hardware: γ-v1 (lockfree) > α (serialized) >
β2-v2 (striped+rwlock). On higher-core-count hardware or
heavier per-op work, the ordering shifts; see
`notes/f32-cache-aware-delivery-plan.md` § F.32-1 for the
per-discipline trade-off table.

### App benches

| Bench | Hale | Go | Node | Python |
|---|---:|---:|---:|---:|
| `stream_aggregator`  | 18.34 ms | 0.23 ms (0.01×) | 1.89 ms (0.10×) | 30.48 ms (1.66×) |

### Refreshing the grid

This grid is a manual snapshot. Refresh it:

- **Before every Hale release.** Bench numbers are part of the
  release evidence — the grid's "Latest snapshot" version + date
  should match the tag being cut.
- **After any substantive runtime / codegen / stdlib change**
  expected to move perf.
- **After any bench-shape change** (new bench file, changed
  iter count, redesigned timed region). The new shape's
  baseline belongs in the grid AND in `baselines.json`.

Regen process:

```sh
# from this dir, with target/release/hale built upstream
./run.sh
```

Copy the per-bench `hale  elapsed_ns=...` + each sibling's
`ratio_vs_hale=...` line into the corresponding row above.
Format times in the most-readable unit (ns / µs / ms / s) and
round to 2-3 significant figures. Update the "Latest snapshot"
line:

- **Version**: matches `../hale/Cargo.toml`'s `[workspace.package]
  version`, the `hale --version` output of the binary used for
  the run, and the git tag being cut. These three should always
  agree at the moment the grid is refreshed.
- **Date**: the date the `./run.sh` was actually executed
  (results/`run-<date>.json` is the canonical artifact).
- **Hardware**: name the CPU + arch + kernel. Numbers shift
  meaningfully between machines (especially L2 size, core
  count, and SMT topology); identifying the host makes the
  grid honest about its scope.

If `./run.sh` regression-fails on a bench, either:
1. Fix the regression (preferred), or
2. Deliberately update `baselines.json` and note WHY in the
   commit message ("baseline drift after F.32-1β2 wired cache-
   padded cells; expected and documented").

Don't refresh the grid against a regressed run without
addressing the regression first — the grid is supposed to
reflect the shipped state of the language, not transient
work-in-progress.

## Three kinds of microbench

The benches under `micro/` split into **three categories**
answering different questions. Mix them with care — they're not
interchangeable.

### Overhead microbenches — "what does the substrate cost when unused?"

Strip allocation work down to nothing so the substrate machinery
runs alone. These deliberately measure Hale's worst case: the
arena lifecycle gets no chance to amortize against work it would
normally accompany.

- **`loop_overhead`** — empty while loop. No arena work at all.
- **`fn_call`** — `fn noop(x) -> Int { return x; }` called 10M
  times. m49's per-call subregion runs against a body that
  doesn't allocate, so the boundary cost is paid for nothing.
- **`locus_instantiation`** — `Empty {}` 100k times,
  statement-position. Arena create + struct init + arena destroy
  with zero allocations in between.
- **`bus_dispatch`** — 10k typed messages through the bus. The
  per-message payload memcpy + queue enqueue is the design (per
  `memory.md` "pointers don't cross loci; values do") but the
  bench measures it isolated.
- **`form_vec_push`** — 500k pushes only. Isolated growth path.
- **`form_vec_get`** — 200k indexed reads only. Isolated read
  path.
- **`form_hashmap_set`** — 1M Int-keyed inserts into a
  `@form(hashmap)` locus. Tests hash + slot probe + entry
  memcpy + occasional grow/rehash.
- **`form_hashmap_get`** — 150k Int-keyed lookups against a
  pre-populated `@form(hashmap)`. Cliffs at n=200k (set+get
  doubles per-iter work vs pure-write set).

Expect Hale to be **slow** here. These benches surface
codegen overhead the compiler team can target (e.g. eliding
arena subregions when a fn provably doesn't allocate) and
violations of spec performance commitments (e.g. `form_vec_get`
60× behind Go violates `spec/forms.md` FORM-3's "within 10%
of C" target).

### Amortized microbenches — "does the design pay off when used as intended?"

Match the shape the design optimizes for: many allocations
inside one arena, wholesale-free at dissolve. The substrate cost
is paid once across N units of work, not per unit.

- **`vec_amortized`** — push N + fold N, single timed region.
- **`fn_scratch_work`** — 100 fn calls × 1000-element local
  `@form(vec)` per call. The m49 subregion gets a real workout.
- **`coord_with_churn`** — chunked-class parent accepting K
  Worker children. Tests F.3 free-list reclamation + chunked
  sub-region allocator. (K capped at 20 by v1 codegen's
  accept() accumulation cliff at k≈25 — see comment in source.)

These are where Hale's region model is supposed to win. The
ratio against Go is the right signal — if amortized benches show
Hale competitive with Go and overhead benches don't, the
design is real but the compiler is leaving per-op cost on the
table.

### Coordinated-workload microbenches — "does the design win in multi-locus shapes?"

The shape Hale is *built for*: multiple loci deep, lateral
siblings, coordinating via vertical-only flow or bus. The design
predicts these should out-throughput dynamic-language alternatives
because region memory + cooperative scheduling + bus-mediated
typed messages avoid the per-allocation GC tracking those
languages pay.

- **`tree_fanout`** — depth=2 with K=20 lateral siblings.
  Coordinator's accept() calls `worker.compute()` for each
  child; results aggregate into parent state. Tests hierarchical
  region memory + cross-locus method dispatch. **First bench
  where Hale decisively beats Node (8.6×) and Python (20.5×).**
- **`pipeline_3stage`** — depth=3 sequential. Source → Filter →
  Sink via two bus subjects. Tests multi-stage bus coordination
  + the cooperative scheduler's drain semantics. Currently
  bottlenecked by per-event bus-dispatch overhead — same shape
  that limits `bus_dispatch`.

### Cross-pool / cache microbenches — "does the substrate respect the cache hierarchy?"

The F.32 cache-aware substrate work (`notes/f32-cache-aware-delivery-plan.md`
in the hale repo) added a new bench family in 2026-05-24 covering
cross-pool dispatch + concurrent `@form(hashmap)` access. These
benches gate the F.32 deliverables — each shipped F.32-N piece
either lands or is rejected against a baseline here.

- **`bus_dispatch_cross_pool`** — same shape as `bus_dispatch`,
  but the subscriber runs on the `io` cooperative pool (a
  separate OS thread). Measures publisher-side enqueue cost
  for the cross-thread mailbox path. Pairs with the prefetch
  hint shipped in F.32-4-prefetch (`__builtin_prefetch` in
  `lotus_coop_pool_post`).
- **`form_hashmap_false_sharing`** — two pools concurrently
  write disjoint keys (even / odd ids) into a shared
  `@form(hashmap, sync = serialized)`. The Prometheus-counter
  shape that drove F.32-0 + F.32-1α. Today's baseline uses
  `sync = serialized` (per-map mutex); when F.32-1β2 ships
  the annotation flips to `sync = striped` for parallel
  writers + cache-padded cells.
- **`form_hashmap_walk_large`** — pre-populate ~100k entries
  then iterate via `key_at` / `entry_at`. TLB-bound at scale;
  used to detect regressions in arena chunk allocation and
  to measure the win from F.32-4a (huge pages, opt-in via
  `LOTUS_HUGE_PAGES=1`).

The cross-pool benches' Go siblings use `runtime.LockOSThread`
to pin goroutines to distinct OS threads — the closest
analogue of "cooperative pool owns its own thread." JS and
Python siblings approximate via `worker_threads` and
`threading.Thread` respectively; both pay overhead that
Hale doesn't (V8 structured-clone, CPython GIL), so the
ratio columns for these benches are informational only,
not strict apples-to-apples.

### App benches — `app/`

- **`stream_aggregator`** — publisher fires N typed events; a
  long-lived aggregator subscribes and maintains running stats.
  Cross between bus_dispatch and a real workload.

## Layout

```
.
├── README.md                this file
├── run.sh                   shell + jq harness
├── baselines.json           checked-in Hale medians + tolerance bands
├── results/                 per-run JSON reports (gitignored)
├── micro/
│   │
│   │   # Overhead microbenches (isolate substrate cost)
│   ├── loop_overhead.{ap,go,js,py}
│   ├── fn_call.{ap,go,js,py}
│   ├── locus_instantiation.{ap,go,js,py}
│   ├── bus_dispatch.{ap,go,js,py}
│   ├── form_vec_push.{ap,go,js,py}
│   ├── form_vec_get.{ap,go,js,py}
│   ├── form_hashmap_set.{ap,go,js,py}
│   ├── form_hashmap_get.{ap,go,js,py}
│   │
│   │   # Amortized microbenches (match design's optimization target)
│   ├── vec_amortized.{ap,go,js,py}
│   ├── fn_scratch_work.{ap,go,js,py}
│   ├── coord_with_churn.{ap,go,js,py}
│   │
│   │   # Coordinated-workload microbenches (multi-locus shapes)
│   ├── tree_fanout.{ap,go,js,py}
│   ├── pipeline_3stage.{ap,go,js,py}
│   │
│   │   # Cross-pool / cache microbenches (F.32)
│   ├── bus_dispatch_cross_pool.{hl,go,js,py}
│   ├── form_hashmap_false_sharing.{hl,go,js,py}
│   └── form_hashmap_walk_large.{hl,go,js,py}
├── app/
│   └── stream_aggregator.{ap,go,js,py}
└── c-twins/                 (placeholder) hand-written C equivalents
                             for FORM-3 10%-gate comparisons.
```

## Conventions

Every bench (any language) self-times the work-of-interest with
a monotonic clock and prints exactly one `elapsed_ns=N` line on
stdout. The harness additionally captures `maxrss_kb` externally
via `/usr/bin/time -v`. The runner takes N samples per bench
(default 5), records the median, and writes both per-sample
arrays and the median into the JSON report.

**Monotonic clocks used:**
- Hale: `std::time::monotonic()` → Duration ns
- Go: `time.Since(t0).Nanoseconds()`
- Node: `process.hrtime.bigint()`
- Python: `time.monotonic_ns()`

**Regression gate (Hale only).** A bench fails when:

```
current_median > baseline_elapsed_ns * (1 + tolerance)
```

Faster-than-baseline is never a regression. The default tolerance
is **0.30** — sub-10ms benches routinely jitter ±20% under OS
noise. Tighten per-bench in `baselines.json` once a metric
stabilizes. Comparative numbers (go/node/python) are emitted in
the report but **never** trigger exit 1.

**Toolchain detection.** The harness checks for `go`, `node`,
and `python3` on PATH at startup. Each comparative language is
silently skipped if its toolchain is missing.

## Adding a bench

1. Decide which category: overhead microbench (isolate a primitive
   under worst-case conditions), amortized microbench (real
   workload shape), or app bench (mixed-workload end-to-end).
2. Drop a `.hl` file in `micro/` or `app/`. Each one
   must:
   - Self-time the work-of-interest with two `std::time::monotonic`
     calls and `t1 - t0` arithmetic.
   - Print `elapsed_ns=` followed by the duration value on its
     own line.
3. Add sibling `.go`, `.js`, `.py` files implementing the same
   shape as closely as the language permits. Each must also
   print one `elapsed_ns=N` line. The harness picks them up
   automatically by filename stem.
4. Run `./run.sh --update-baselines` to seed.
5. Commit the new sources + the updated `baselines.json`.

## Reading the report

```json
{
  "generated_at": "...",
  "iters": 5,
  "benches": [
    {
      "name": "fn_scratch_work",
      "kind": "micro",
      "status": "ok",
      "elapsed_ns_median": 469208,
      "elapsed_ns_samples": [...],
      "maxrss_kb_median": 3088,
      "maxrss_kb_samples": [...],
      "baseline_elapsed_ns": 469208,
      "baseline_maxrss_kb": 3088,
      "note": null,
      "comparatives": {
        "go":     { "elapsed_ns_median": 397163,  "ratio_vs_hale": 0.8465, ... },
        "node":   { "elapsed_ns_median": 918277,  "ratio_vs_hale": 1.9571, ... },
        "python": { "elapsed_ns_median": 2578210, "ratio_vs_hale": 5.4948, ... }
      }
    }
  ]
}
```

`ratio_vs_hale = lang_elapsed / hale_elapsed`. A value of
**0.5** means the other language is 2× faster than Hale; **2.0**
means Hale is 2× faster than the other language; **1.0** is
parity.

## Known constraint — accumulation ceiling

Several Hale substrate paths segfault under v1 codegen
somewhere between 100k and 1M iterations of a tight loop. The
microbench Hale iteration counts are tuned **below** those
ceilings; comments inside each `.hl` document the threshold
they hit. The sibling `.go/.js/.py` benches mirror the same
iteration count so the ratio stays apples-to-apples.

The **`coord_with_churn`** bench hits the steepest cliff:
parent's `accept(child)` in a loop fails at k≈25 regardless of
projection class. Caps the bench at K=20, where the timing
signal is small but measurable. When the substrate
accumulation fix lands, raise iteration counts in all four
language files together.

## Future work

- **C twins for FORM-3.** `spec/forms.md` commits `@form(vec)`
  to within 10% of a hand-written C equivalent on push+get.
  Land C sources in `c-twins/` and add a comparison column
  to the runner's report.
- **More comparative langs.** Erlang (BEAM) is the natural fourth
  comparator since Hale's runtime model is BEAM-shaped. Rust
  is a fifth for the "non-GC compiled" point of comparison.
- **`hale bench` CLI.** Per `spec/testing.md`, this surface is
  planned but not shipped. The shell harness here is the current
  stand-in.
