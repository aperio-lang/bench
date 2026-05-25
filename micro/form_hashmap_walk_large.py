"""Python equivalent of form_hashmap_walk_large.hl.

Pre-populate a dict, iterate via .items(). Sum the values to
defeat dead-code elimination. dict iteration in CPython is
insertion-ordered (since 3.7) and the per-entry cost includes
Python's per-iteration interpreter overhead.
"""

import time

n = 100_000
m = {i: (i, i * 3) for i in range(n)}

t0 = time.monotonic_ns()
sum_ = 0
for k, (_id, v) in m.items():
    sum_ += v + (k - k)
elapsed = time.monotonic_ns() - t0

print(f"n={n}")
print(f"len={len(m)}")
print(f"sum={sum_}")
print(f"elapsed_ns={elapsed}")
