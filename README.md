# fencelock

A Go distributed lock library built around **fencing tokens**: every lock
acquisition hands out a monotonically increasing token, and the protected
resource rejects any write that arrives with a stale token. This turns
"mutual exclusion" (unenforceable across a network) into "ordered writes
with staleness rejection" (enforceable).

Backends: **etcd** (primary, linearizable) and **Redis** (secondary,
best-effort — kept intentionally to demonstrate where it's weaker).

> Status: design phase. See [docs/ROADMAP.md](docs/ROADMAP.md) for what's
> built vs. planned.

## Documentation

| Page | What it covers |
|---|---|
| [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) | Functional and non-functional requirements, explicit non-goals, and the test-to-requirement traceability table. Start here to understand *what* this library must do and how "done" is defined. |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Phased build plan with live status tracking, from skeleton through publishing the write-up. Start here to see what's next. |
| [docs/adr/](docs/adr) | Architecture Decision Records — the *why* behind key choices: etcd-vs-Redis, fencing token source of truth, server-side TTL enforcement, and `FencedResource` as a first-class API. Read these when a design choice looks surprising. |

## Quickstart

```bash
go test ./...
```

(Full instructions land once Phase 1 of the roadmap is complete — see
[docs/ROADMAP.md](docs/ROADMAP.md).)

## Core idea

A distributed lock (Redis `SETNX`, etcd lease, ZooKeeper znode) can never
*guarantee* mutual exclusion: a lock holder can pause (GC, VM stall, network
partition) longer than its lease TTL, the lock expires, another client
acquires it, and now two clients believe they hold the lock at the same
time. You can't fix this with better timeouts. The fix: every time the lock
is granted, hand out a monotonically increasing fencing token, and require
the protected resource to reject any write carrying a token lower than the
highest it has already seen.

See [docs/REQUIREMENTS.md](docs/REQUIREMENTS.md) §0 for the full framing,
and [docs/adr/0002-fencing-token-source-of-truth.md](docs/adr/0002-fencing-token-source-of-truth.md)
for how the token itself is sourced and why it must never be client-computed.
