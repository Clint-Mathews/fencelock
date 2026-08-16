# ADR 0004: FencedResource is a first-class interface, not a test-only helper

## Status
Accepted

## Context
Most naive lock-library demos stop at "the lock API returns a token" and
never show enforcement — but that's the part that actually matters. If
`FencedResource` were just an internal test fixture rather than
a designed, exported interface, the project would risk making exactly that
mistake — proving the token exists without proving it does anything.

## Decision
- `FencedResource` is defined in `lock/` alongside `Locker`/`Lease`, as a
  public interface: `Write(ctx, token int64, data []byte) error`.
- At least one concrete implementation (`fencedstore.Memory`) ships as part
  of the library, not just inside test code, so downstream users have a
  working reference for wiring their own resource up to token checks.
- A second implementation (`fencedstore.Postgres`, `UPDATE ... WHERE
  last_token < $1`) ships to demonstrate the pattern generalizes past an
  in-memory toy to something resembling a real datastore.
- The test suite's Test A/B pair is built specifically to make the
  enforcement visible: Test A shows corruption without the check, Test B
  shows rejection with it, using the same race scenario.

## Consequences
- `fencedstore` depends only on `lock`, never on `etcdlock`/`redislock` —
  enforcement must be demonstrably backend-agnostic, since the whole point
  is that the resource, not the lock service, is what enforces ordering.
- The README/blog explicitly documents the limitation this implies: fencing
  tokens only work when the resource cooperates. You cannot fence a write to
  an arbitrary third-party API that doesn't check tokens — this is not a bug
  to fix, it's a boundary of the pattern, and it must be stated honestly
  (see REQUIREMENTS.md NFR-6).
