// JS equivalent of form_hashmap_false_sharing.hl.
//
// Node's worker_threads can't share a JS Map across workers —
// the closest analogue is `SharedArrayBuffer` with Atomics for
// integer slots, but a structured shared Map isn't part of the
// API. Instead, this benchmark spawns a worker that owns the
// map and the main thread posts mutate messages over a
// MessageChannel. The handoff is serialized (single owner),
// which matches the Hale `sync = serialized` shape — but the
// cost shape is dominated by structured-clone serialization
// (Node's postMessage path), not by mutex contention.
//
// Expect this to be slower than Hale by 1-2 orders of
// magnitude on this workload — that's the honest cost of
// JS's cross-thread story for shared mutable state.

const { Worker, isMainThread, parentPort } = require('worker_threads');

if (isMainThread) {
    const perWriter = 100000;
    const total = perWriter * 2;
    const worker = new Worker(__filename);

    let elapsed = 0n;

    worker.on('message', (msg) => {
        if (msg.type === 'len') {
            console.log(`per_writer=${perWriter}`);
            console.log(`total=${msg.value}`);
            console.log(`elapsed_ns=${elapsed}`);
            worker.terminate();
        }
    });

    const t0 = process.hrtime.bigint();

    // Producer A — even ids
    for (let i = 0; i < perWriter; i++) {
        worker.postMessage({ type: 'set', id: i * 2, v: i });
    }
    // Producer B — odd ids (interleaved is closer to true
    // adversarial false-sharing; sequential per-writer also OK
    // here because worker drains serially regardless)
    for (let i = 0; i < perWriter; i++) {
        worker.postMessage({ type: 'set', id: i * 2 + 1, v: i });
    }

    worker.postMessage({ type: 'done' });
    elapsed = process.hrtime.bigint() - t0;
} else {
    const entries = new Map();
    parentPort.on('message', (msg) => {
        if (msg.type === 'set') {
            entries.set(msg.id, { id: msg.id, v: msg.v });
        } else if (msg.type === 'done') {
            parentPort.postMessage({ type: 'len', value: entries.size });
        }
    });
}
