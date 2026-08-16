# Requirements

This document lays out testable functional and non-functional requirements
so scope and "done" are unambiguous.

## 0. Goal

Prove, end-to-end, that fencing tokens turn an unenforceable distributed
mutual-exclusion problem ("only one client holds the lock") into an
enforceable ordered-writes problem ("the resource rejects stale writes"),
via a working Go library plus a test suite that reproduces the failure and
the fix.

Success is **not** "a lock API that returns a token." Success is a resource
that demonstrably rejects a stale write from a client that legitimately once
held the lock.

## 1. Functional Requirements

### FR-1 — Locker interface
- FR-1.1: `Acquire(ctx, resource, ttl) (*Lease, error)` blocks until acquired or `ctx` is cancelled.
- FR-1.2: `TryAcquire(ctx, resource, ttl) (*Lease, error)` returns immediately (non-blocking) if the lock is held elsewhere.
- FR-1.3: `Lease` carries `Resource string`, `Token int64`, `ExpiresAt time.Time`.
- FR-1.4: `Lease.Release(ctx) error` releases the lock early.
- FR-1.5: `Lease.Refresh(ctx, ttl) error` heartbeats/renews the lease.
- FR-1.6: `Lease.Valid() bool` reports local TTL expiry (best-effort, client-side only — not a substitute for server-side enforcement).

### FR-2 — Fencing tokens
- FR-2.1: Every successful `Acquire`/`TryAcquire` returns a token strictly greater than any previously issued token for that `resource`.
- FR-2.2: Two concurrent acquire calls for the same resource never produce the same token.
- FR-2.3: Token source must be server-side and monotonic (etcd: mod revision; Redis: atomic `INCR`) — never client-computed.

### FR-3 — FencedResource (resource-side enforcement)
- FR-3.1: `Write(ctx, token, data) error` succeeds only if `token >= last-seen token` for that resource.
- FR-3.2: A stale write (token lower than the highest seen) is rejected with a distinguishable error, and must **not** mutate state.
- FR-3.3: At least one in-memory implementation (`fencedstore.Memory`) must exist and be race-safe under `-race`.
- FR-3.4: A second implementation backed by Postgres (`UPDATE ... WHERE last_token < $1`) is required to show the pattern isn't memory-only.

### FR-4 — Backends
- FR-4.1: An etcd-backed `Locker` implementation is required (primary backend), using `clientv3/concurrency` sessions/mutexes with the revision number exposed as the fencing token.
- FR-4.2: A Redis-backed `Locker` implementation is required (secondary backend), using `SET NX PX` + Lua-scripted compare-and-delete release + `INCR` for token issuance.
- FR-4.3: Both backends implement the same `Locker` interface from `lock/`; no backend-specific types leak into `lock/`.
- FR-4.4: The Redis backend must be documented (README + blog) as "best-effort," explicitly citing the Redlock critique — it must not be presented as equivalent-strength to etcd.

### FR-5 — Demo CLI
- FR-5.1: `cmd/demo` must let a user run two processes and visually observe: client A acquires → client A "pauses" past TTL → client B acquires with a higher token → client A's write is rejected by the fenced resource.

## 2. Non-Functional Requirements

- NFR-1 (Correctness under concurrency): `go test -race ./...` must pass; the concurrency stress test (N goroutines racing to acquire the same lock repeatedly) must show strictly increasing tokens and no overlapping lease windows.
- NFR-2 (No mocking of distributed behavior): Tests exercising real race/partition conditions must run against real etcd/Redis via `testcontainers-go`, not mocks — mocking the race defeats the purpose of the project.
- NFR-3 (Server-side TTL): Lease expiry enforcement must not depend solely on wall-clock comparisons computed by the client; etcd lease TTL is enforced server-side.
- NFR-4 (API clarity): The `Locker` and `FencedResource` interfaces must be defined and stable before backend implementation begins (interface-first, per Phase 1).
- NFR-5 (CI): GitHub Actions must run the full test suite, including containerized-backend tests, on every push/PR.
- NFR-6 (Honesty in docs): README/blog must explicitly state what fencing tokens do *not* solve (can't fence a write to a third-party API that doesn't check tokens) rather than overclaiming.

## 3. Explicit Non-Goals

- Not building a general-purpose distributed coordination service (that's etcd/Zookeeper's job).
- Not attempting to make single-instance Redis linearizable — the point is to *show* where it's weaker, not fix it.
- Not fencing arbitrary third-party APIs that don't cooperate with token checks (out of scope by definition — documented as a known limitation).

## 4. Test Requirements (traceable to FRs)

| Test | Proves | Requirement(s) |
|---|---|---|
| Test A — naive lock fails | Unguarded write after TTL-expiry pause corrupts state | Motivates FR-3 |
| Test B — fencing token saves it | Same scenario through `FencedResource.Write` rejects the stale write | FR-3.1, FR-3.2 |
| Test C — concurrency stress | Strictly increasing tokens, no overlapping leases, `-race` clean | FR-2.1, FR-2.2, NFR-1 |
| Test D — etcd vs Redis comparison | Same partition simulation on both backends; documents where Redis can still admit a hazard etcd doesn't | FR-4.4 |

## 5. Deliverables Outside Code

- README with an architecture/sequence diagram of the pause → expire → fence scenario.
- Blog post: "Why your distributed lock is probably broken."
- LinkedIn post + X/Twitter thread reusing the same diagram.
- Publish order: repo + tests + README → blog → LinkedIn → X thread, spaced a day or two apart.
