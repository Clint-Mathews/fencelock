# ADR 0003: Lease TTL/expiry must be enforced server-side, not by client wall-clock comparison

## Status
Accepted

## Context
Kleppmann's core scenario is a client that
pauses — GC, VM stall, network partition — for longer than its lease TTL. If
expiry were only a client-side computation ("is `now() > expiresAt`?"), the
paused client's own clock cannot be trusted to have advanced correctly, and
worse, a slow-but-not-paused client could simply be wrong about whether its
lease has expired due to clock drift between it and the lock service.

The fix the project is built around (fencing tokens) is what makes the
*consequence* of this problem harmless. But that doesn't mean expiry itself
should be left to client-side clocks where a server-side mechanism is
available — sloppy expiry just means a lock gets contended/released
inconsistently, which is separately worth avoiding.

## Decision
- etcd backend: rely on etcd's lease TTL, enforced server-side by etcd
  itself (the lease is revoked by the etcd cluster on expiry, independent of
  what any client believes about elapsed time).
- Redis backend: rely on Redis's own `PX` expiry on the lock key, enforced
  server-side by Redis, not by a client computing "should this have expired
  by now."
- `Lease.Valid()` remains a client-side, best-effort convenience check only
  (e.g., to short-circuit before attempting a doomed `Refresh`), and must be
  documented as such — it is never the thing a caller should rely on for
  correctness.

## Consequences
- Correctness of "is the lock still held" never depends on comparing two
  different machines' clocks.
- This ADR is orthogonal to fencing tokens: even with server-side TTL
  enforcement, a client can still act after its lease has server-side
  expired (that's exactly the scenario the tests reproduce) — which is why
  ADR 0002 and this ADR are both required, not either/or.
