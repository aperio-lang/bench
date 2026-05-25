// JS equivalent of form_hashmap_walk_large.hl.
//
// Pre-populate a Map<int, {id, val}>, iterate via Map.forEach.
// V8's Map is a hash table (insertion-ordered); iteration cost
// per entry includes a function call overhead from forEach.
// Equivalent shape to Hale's key_at/entry_at sweep.

const n = 100_000;
const m = new Map();
for (let i = 0; i < n; i++) {
    m.set(i, { id: i, val: i * 3 });
}

const t0 = process.hrtime.bigint();
let sum = 0;
m.forEach((e, k) => {
    sum += e.val + (k - k);
});
const elapsed = process.hrtime.bigint() - t0;

console.log(`n=${n}`);
console.log(`len=${m.size}`);
console.log(`sum=${sum}`);
console.log(`elapsed_ns=${elapsed}`);
