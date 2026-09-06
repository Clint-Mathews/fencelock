# Roadmap

Phase status only. Requirements, non-goals, and the test-to-requirement
map live in [REQUIREMENTS.md](REQUIREMENTS.md). Design rationale lives in
[adr/](adr). Do not copy those texts here — link to the FR/NFR or file.

## Phase 1 — Skeleton
- [x] Module and packages `lock/`, `etcdlock/`, `redislock/`, `fencedstore/`, `cmd/demo/`
- [x] Interfaces in `lock/` ([FR-1](REQUIREMENTS.md), [FR-3](REQUIREMENTS.md), [NFR-4](REQUIREMENTS.md))

## Phase 2 — etcd implementation
- [x] `etcdlock/` ([FR-4.1](REQUIREMENTS.md), [ADR 0002](adr/0002-fencing-token-source-of-truth.md), [ADR 0003](adr/0003-server-side-ttl-enforcement.md))
- [x] `docker-compose.yml` — local etcd (and Redis) for **demo / manual** use, not for `go test`

## Phase 3 — Redis implementation
- [x] `redislock/` ([FR-4.2](REQUIREMENTS.md), [ADR 0001](adr/0001-backend-etcd-primary-redis-secondary.md))
- [x] README documents Redis as weaker / Redlock counterexample ([FR-4.4](REQUIREMENTS.md))

## Phase 4 — Toy fenced resource
- [x] `fencedstore.Memory` ([FR-3.3](REQUIREMENTS.md))
- [x] `fencedstore.Postgres` ([FR-3.4](REQUIREMENTS.md))

## Phase 5 — Test suite
- [x] Test A — [`test/integration/naive_lock_fails_test.go`](../test/integration/naive_lock_fails_test.go)
- [x] Test B — [`test/integration/fencing_saves_it_test.go`](../test/integration/fencing_saves_it_test.go)
- [x] Test C — [`test/integration/concurrency_stress_test.go`](../test/integration/concurrency_stress_test.go)
- [x] Test D — [`test/integration/etcd_vs_redis_test.go`](../test/integration/etcd_vs_redis_test.go)
- [x] Real etcd/Redis via testcontainers ([NFR-2](REQUIREMENTS.md)); mapping of tests to FRs is in [REQUIREMENTS.md §4](REQUIREMENTS.md)

## Phase 6 — Polish
- [x] README sequence diagram (pause → expire → fence) ([§5](REQUIREMENTS.md))
- [x] `cmd/demo` ([FR-5](REQUIREMENTS.md)) — single process against compose etcd; see README
- [x] GitHub Actions [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) ([NFR-5](REQUIREMENTS.md))

## Explicitly deferred / stretch (not required for v1)

Listed as non-goals / stretch in [REQUIREMENTS.md](REQUIREMENTS.md):

- Postgres advisory-lock backend as a third `Locker`
- Any backend beyond etcd + Redis

## Related docs

- [REQUIREMENTS.md](REQUIREMENTS.md) — what “done” means
- [adr/](adr) — why the backends and token design are the way they are
- [README.md](../README.md) — how a stranger builds, tests, and runs the demo
