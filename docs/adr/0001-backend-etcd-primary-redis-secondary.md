# ADR 0001: etcd as primary backend, Redis as secondary/best-effort backend

## Status
Accepted

## Context
A distributed lock needs a backend that hands out leases and, ideally, a
monotonic counter to use as a fencing token. Two realistic options:

- **Redis (single instance)**: fast to build, but inherits the Redlock
  critique directly — Kleppmann's argument is that a single-node (or even
  multi-node Redlock) design cannot guarantee the mutual-exclusion property
  it appears to provide, because it has no linearizable source of truth for
  time or ordering. Antirez's rebuttal disputes the practical severity but
  does not dispute the theoretical gap.
- **etcd**: Raft-backed, linearizable reads/writes, and every write already
  carries a monotonically increasing revision number per key — which is
  functionally identical to the "sequencer" concept from Google's Chubby
  paper (Chang et al.), predating Kleppmann's post by roughly a decade.

Building against only Redis would make the project's central claim (fencing
tokens close a real correctness gap) harder to prove convincingly, since
Redis itself is the weaker foundation the post is arguing against. Building
against only etcd would lose the chance to demonstrate, concretely, *why*
Redis is weaker — an assertion is not as strong as a reproducible failure.

## Decision
Implement etcd as the primary backend. Its `concurrency.Mutex`/session
primitives supply the lock; the created key's mod revision supplies the
fencing token, so correctness of the counter is inherited from etcd's Raft
consensus rather than re-implemented.

Implement Redis as a secondary backend behind the same `Locker` interface,
explicitly labeled "best-effort." Use it in the test suite (Test D) and the
writeup to demonstrate the specific hazard etcd's linearizability closes.

## Consequences
- `lock/` interfaces must not assume anything etcd-specific (e.g., no
  leaking `clientv3` types), so the Redis implementation can satisfy the
  same contract.
- The Redis implementation still needs its own fencing counter (`INCR` on a
  separate key), since Redis has no built-in per-key revision.
- Documentation (README, blog) must state the strength difference
  explicitly rather than presenting both backends as equivalent — this is a
  requirement (see REQUIREMENTS.md FR-4.4), not just a nicety.
- Test D (etcd vs Redis comparison) becomes required, not optional, because
  it's the only place the claimed hazard is actually demonstrated rather
  than asserted.

## References
- Martin Kleppmann, "How to do distributed locking" (2016)
- Kleppmann's Redlock critique and Antirez's rebuttal
- Chang et al., "The Chubby lock service for loosely-coupled distributed systems"
