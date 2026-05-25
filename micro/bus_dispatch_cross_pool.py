"""Python equivalent of bus_dispatch_cross_pool.hl.

threading.Thread + queue.Queue is the closest analogue of
"cooperative(pool = io) subscriber on a separate scheduler" —
both sides are real threads, the Queue handles handoff.

Caveat: Python's GIL means only one thread executes Python
bytecode at a time, so the producer and consumer don't actually
run in parallel. Queue.put/get release the GIL around the
underlying mutex, so the handoff is real, but the per-iteration
overhead is dominated by GIL acquire/release plus Lock contention.
This is the honest cost of cross-thread message-passing in
Python; multiprocessing would add pickle serialization on top.
"""

import queue
import threading
import time


class Tick:
    __slots__ = ("n",)
    def __init__(self, n): self.n = n


def consumer(q, done):
    count = 0
    while True:
        msg = q.get()
        if msg is None:
            done.append(count)
            return
        count += 1


iters = 100_000
q = queue.Queue(maxsize=64)
done = []
t = threading.Thread(target=consumer, args=(q, done))
t.start()

t0 = time.monotonic_ns()
for i in range(iters):
    q.put(Tick(i))
elapsed = time.monotonic_ns() - t0

q.put(None)
t.join()

print(f"iters={iters}")
print(f"count={done[0]}")
print(f"elapsed_ns={elapsed}")
