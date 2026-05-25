"""Python equivalent of form_hashmap_false_sharing.hl.

Two threads hammer disjoint keys of a shared dict. The GIL
serializes Python bytecode execution, so threads don't run in
parallel — but a threading.Lock-protected dict is the closest
analogue of Hale's `sync = serialized` shape: explicit per-
map lock, sequential mutations. The bench measures dict-set +
lock-acquire cost; expected throughput is significantly lower
than Hale or Go because Python's per-op overhead is heavy
even before adding sync.
"""

import threading
import time


class Registry:
    __slots__ = ("lock", "entries")
    def __init__(self):
        self.lock = threading.Lock()
        self.entries = {}

    def set(self, id_, v):
        with self.lock:
            self.entries[id_] = (id_, v)

    def __len__(self):
        with self.lock:
            return len(self.entries)


per_writer = 100_000
total_expected = per_writer * 2
reg = Registry()

def writer_a():
    for i in range(per_writer):
        reg.set(i * 2, i)

def writer_b():
    for i in range(per_writer):
        reg.set(i * 2 + 1, i)

t_a = threading.Thread(target=writer_a)
t_b = threading.Thread(target=writer_b)

t0 = time.monotonic_ns()
t_a.start()
t_b.start()
t_a.join()
t_b.join()
elapsed = time.monotonic_ns() - t0

print(f"per_writer={per_writer}")
print(f"total={len(reg)}")
print(f"elapsed_ns={elapsed}")
