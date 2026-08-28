# Roadmap

Phased build plan (~8–16 hrs total). Status is tracked here as the single
source of truth for progress — update it as work lands instead of writing
separate status notes.

## Phase 1 — Skeleton (1–2 hrs)
- [x] `go mod init`
- [x] Create `lock/`, `etcdlock/`, `redislock/`, `fencedstore/`, `cmd/demo/` per [STRUCTURE.md](STRUCTURE.md)
- [x] Define `Locker`, `Lease`, `FencedResource` interfaces in `lock/` (no implementations yet)

## Phase 2 — etcd implementation (2–4 hrs)
- ☐ Wire up `go.etcd.io/etcd/client/v3/concurrency` sessions/mutexes
- ☐ Expose the created key's mod revision as `Lease.Token`
- ☐ Implement `Acquire` / `TryAcquire` / `Release` / `Refresh`
- ☐ `docker-compose.yml` with local etcd for dev/test

## Phase 3 — Redis implementation (2–3 hrs, optional but recommended)
- ☐ `SET key value NX PX ttl` acquire
- ☐ Lua script for safe compare-and-delete release (client-owned token)
- ☐ `INCR` on separate counter key for fencing token issuance
- ☐ Document this backend as intentionally "weaker" — the Redlock counterexample

## Phase 4 — Toy fenced resource (1–2 hrs)
- ☐ `fencedstore.Memory`: `map[string]int64` last-seen-token, race-safe
- ☐ `fencedstore.Postgres`: `UPDATE ... WHERE last_token < $1`

## Phase 5 — Test suite (2–4 hrs — highest-signal phase)
- ☐ Test A: naive lock fails (unguarded write after pause corrupts state)
- ☐ Test B: fencing token saves it (same scenario via `FencedResource.Write`)
- ☐ Test C: concurrency stress (N goroutines, strictly increasing tokens, `-race`)
- ☐ Test D: etcd vs Redis comparison (same partition sim, document the hazard gap)
- ☐ `testcontainers-go` for real etcd/Redis in CI (no mocking distributed races)

## Phase 6 — Polish (2–3 hrs)
- ☐ README with architecture/sequence diagram (pause → expire → fence)
- ☐ `cmd/demo`: two-process CLI demo of the race + rejection, live
- ☐ GitHub Actions CI running full suite incl. containerized backends

## Phase 7 — Publish (spaced a day or two apart)
- ☐ Repo + tests + solid README live, `go test ./...` runnable by strangers
- ☐ Blog post: "Why your distributed lock is probably broken"
- ☐ LinkedIn post (reuse sequence diagram, link in first comment)
- ☐ X/Twitter thread (hook → diagram → insight → code → repo link → blog link → source references)
- ☐ Follow-up single-tweet repost a day or two later

## Explicitly deferred / stretch (not required for v1)
- Postgres advisory-lock backend as a third `Locker` implementation (mentioned as a possible extension in the blog CTA).
- Any backend beyond etcd + Redis.

## Related docs
- [STRUCTURE.md](STRUCTURE.md) — target repo layout
- [REQUIREMENTS.md](REQUIREMENTS.md) — functional/non-functional requirements this roadmap implements
- [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md) — code-level, phase-by-phase build instructions for everything above
- [adr/](adr) — architecture decisions made along the way
