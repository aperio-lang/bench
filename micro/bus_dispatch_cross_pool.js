// JS equivalent of bus_dispatch_cross_pool.hl.
//
// worker_threads is the canonical cross-thread primitive in
// Node. postMessage structurally clones the payload — so this
// bench includes serialization cost Hale's typed bus does NOT
// pay (Hale memcpys a flat struct; Node serializes per V8's
// structured-clone algorithm). Expect Node to be substantially
// slower than Hale here; that's the true cost of "message
// crosses a thread boundary in Node".

const { Worker, isMainThread, parentPort } = require('worker_threads');

if (isMainThread) {
    const iters = 100000;
    const worker = new Worker(__filename);

    let count = 0;
    let elapsed = 0n;

    worker.on('message', (msg) => {
        if (msg.type === 'count') {
            count = msg.value;
            console.log(`iters=${iters}`);
            console.log(`count=${count}`);
            console.log(`elapsed_ns=${elapsed}`);
            worker.terminate();
        }
    });

    const t0 = process.hrtime.bigint();
    for (let i = 0; i < iters; i++) {
        worker.postMessage({ type: 'tick', n: i });
    }
    elapsed = process.hrtime.bigint() - t0;
    worker.postMessage({ type: 'done' });
} else {
    let count = 0;
    parentPort.on('message', (msg) => {
        if (msg.type === 'tick') {
            count++;
        } else if (msg.type === 'done') {
            parentPort.postMessage({ type: 'count', value: count });
        }
    });
}
