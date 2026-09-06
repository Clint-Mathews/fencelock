# Roadmap

Phased build plan (~8–16 hrs total). Status is tracked here as the single
source of truth for progress — update it as work lands instead of writing
separate status notes.

## Phase 1 — Skeleton
- [x] `go mod init`
- [x] Create `lock/`, `etcdlock/`, `redislock/`, `fencedstore/`, `cmd/demo/` per [STRUCTURE.md](STRUCTURE.md)
- [x] Define `Locker`, `Lease`, `FencedResource` interfaces in `lock/` (no implementations yet)

## Phase 2 — etcd implementation
- [x] Wire up `go.etcd.io/etcd/client/v3/concurrency` sessions/mutexes
- [x] Expose the created key's mod revision as `Lease.Token`
- [x] Implement `Acquire` / `TryAcquire` / `Release` / `Refresh`
- [x] `docker-compose.yml` with local etcd for dev/test

## Phase 3 — Redis implementation
- [x] `SET key value NX PX ttl` acquire
- [x] Lua script for safe compare-and-delete release (client-owned token)
- [x] `INCR` on separate counter key for fencing token issuance
- [x] Document this backend as intentionally "weaker" — the Redlock counterexample

## Phase 4 — Toy fenced resource
- [x] `fencedstore.Memory`: `map[string]int64` last-seen-token, race-safe
- [x] `fencedstore.Postgres`: `UPDATE ... WHERE last_token < $1`

## Phase 5 — Test suite
- [x] Test A: naive lock fails (unguarded write after pause corrupts state)
- [x] Test B: fencing token saves it (same scenario via `FencedResource.Write`)
- [x] Test C: concurrency stress (N goroutines, strictly increasing tokens, `-race`)
- [x] Test D: etcd vs Redis comparison (same partition sim, document the hazard gap)
- [x] `testcontainers-go` for real etcd/Redis in CI (no mocking distributed races)

## Phase 6 — Polish
- [x] README with architecture/sequence diagram (pause → expire → fence)
- [x] cmd/demo`: two-process CLI demo of the race + rejection, live
- [x] GitHub Actions CI running full suite incl. containerized backends

## Explicitly deferred / stretch (not required for v1)
- Postgres advisory-lock backend as a third `Locker` implementation (mentioned as a possible extension in the blog CTA).
- Any backend beyond etcd + Redis.

## Related docs
- [STRUCTURE.md](STRUCTURE.md) — target repo layout
- [REQUIREMENTS.md](REQUIREMENTS.md) — functional/non-functional requirements this roadmap implements
- [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) — code-level, phase-by-phase build instructions for everything above
- [adr/](adr) — architecture decisions made along the way
