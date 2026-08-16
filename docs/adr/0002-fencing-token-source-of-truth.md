# ADR 0002: Fencing token must be a server-side monotonic value, never client-computed

## Status
Accepted

## Context
The entire value of this project rests on the fencing token being a
trustworthy total order. If the token were computed client-side (e.g., a
local counter, or a timestamp), a paused or clock-skewed client could
produce a token that looks valid but isn't actually ordered correctly
relative to other clients — reintroducing exactly the hazard the project
exists to eliminate (see the clock-drift failure mode discussed in
REQUIREMENTS.md).

Two backends, two different natural sources of a monotonic value:
- etcd: every key write has a `Revision`/mod-revision that is strictly
  increasing per-key, assigned by Raft consensus.
- Redis: no built-in per-key revision; the natural mechanism is `INCR` on a
  dedicated counter key, which is atomic given Redis's single-threaded
  command execution.

## Decision
- etcd backend: expose the mod revision of the lock key as `Lease.Token`.
  Do not derive the token from wall-clock time or a client-side counter.
- Redis backend: issue tokens via `INCR` on a counter key scoped per
  resource, atomically with (or immediately after) the `SET NX PX` that
  acquires the lock.
- In both backends, `Lease.Valid()` (client-side TTL check) is documented as
  advisory only — it exists for the client to avoid wasted work, not as the
  mechanism that makes writes safe. The `FencedResource.Write` check against
  the token is what makes writes safe.

## Consequences
- The `Locker` interface must return the token as part of `Lease` on every
  successful acquire, not as a side channel — callers must not be able to
  "forget" to fetch it.
- `FencedResource` implementations must compare tokens with a total order
  (`int64`, `>=`) and must reject on tie-or-lower, never attempt to
  reconcile or merge concurrent writes.
- This decision is what Test C (concurrency stress) and Test B (fencing
  saves it) actually verify — "tokens strictly increasing" and "stale token
  rejected" are the direct, testable consequences of this ADR.

## References
- Chubby's sequencer concept (Chang et al.)
- REQUIREMENTS.md — "Failure modes to explicitly design for": clock drift
