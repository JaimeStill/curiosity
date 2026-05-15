# ECS

The Entity-Component-System substrate. Inner-tier; its storage
layout is the substrate every other inner-tier member reads from or
writes to. This document is forward-looking reference material —
claims grounded in source code or in hard external constraints.
Material that is not yet settled lives under `concepts/engine/`,
not here.

## Position in the engine

Package root: `engine/ecs/`.

Sub-packages, applying the workspace's per-primitive package
pattern at the engine layer:

- `engine/ecs/entity/` — entity-identity primitives.
- `engine/ecs/component/` — component-identity primitives.
- `engine/ecs/archetype/` — archetype storage implementation
  (the default backend; the typed call-site surface is a direct
  facade over archetype with no internal type-erased layer).
- `engine/ecs/system/` — *deferred per iterative depth*. System
  primitives remain concept-tier in `concepts/engine/scheduler.md`;
  the package is introduced when those primitives firm up.

The `engine/ecs/` root package itself holds the `World` type and
the typed call-site surface (Spawn, Despawn, Attach, Detach, Get,
Set, Has, NewQuery). Each `*World` is concretely backed by
archetype data structures; there is no abstract `Storage` interface
in the engine's hot path.

Dependency direction inside `engine/ecs/` is strictly one-way:
`entity/` and `component/` are leaf primitives that import only
the standard library; `archetype/` imports `entity/` and
`component/`; the root package imports all three. No cycles.

## Sub-package responsibilities

### `engine/ecs/entity/`

Owns the entity-identity surface. `entity.ID` and `entity.Allocator`
are implemented; their shape, helpers, and method contracts live in
`engine/ecs/entity/entity.go` and `engine/ecs/entity/allocator.go`
plus their godoc.

Forward-looking: each `*World` instance will own its own allocator
as an unexported field. The world's Spawn / Despawn methods go
through the allocator's `Allocate` / `Recycle` surface; every API
call taking an `entity.ID` calls `Allocator.Validate` and
translates a `false` result into `ErrStaleEntity` at the world
boundary.

### `engine/ecs/component/`

Owns the component-identity surface. `component.ID`,
`component.Value`, the type→ID registry (`IDFor[T]` / `TypeOf`),
and `component.Signature` plus its methods are implemented across
`engine/ecs/component/component.go`,
`engine/ecs/component/registry.go`, and
`engine/ecs/component/signature.go` with godoc on every exported
symbol.

`MaxCID = 2048` is the upper bound on simultaneously-registered
component types within one engine instance. The constant sizes
both the archetype storage's per-CID column index (a flat array)
and the `Signature` bitset (`[MaxCID/64]uint64` — 32 uint64 words,
~256 B per signature; archetype's `columnFor` at ~4 KB per
archetype). Aligning `Signature` and `columnFor` against the same
constant is deliberate: archetype signatures compare bit-for-bit
without dynamic width handling, and downstream code has a single
point to revisit if the bound changes. The plan-exit alternative —
carrying a uint64 signature forward and revisiting at archetype
time — was rejected because it would force a separate width
invariant for `Signature` vs. `columnFor` and risk drift between
them. The 2048 value pays a modest per-archetype memory cost and
is well past plausible component-type counts; revisit if the
"archetype-table memory growth past 100k entities" open question
(see Forward-looking below) materializes a budget problem against
the per-archetype ~4 KB column index.

Forward-looking: archetype keys signatures, queries match against
them, and the typed API surface derives CIDs via `IDFor[T]()`.
Each integration constraint lands when its consumer lands.

### `engine/ecs/archetype/`

Owns archetype storage. Implements the storage substrate that
won the cross-backend comparison (see
`experiments/ecs-storage-comparison/`'s Finding).

Each archetype is a column-store keyed by component signature.
Locations are tracked in an ID-indexed `locations` slice (entity
index → archetype reference + row); the slice replaces an earlier
map-based representation that surfaced as a heap and per-spawn
hotspot at scale. Each archetype maintains a flat-array column
index `columnFor [MaxCID]int16` (column index per CID; `-1` if
absent) to avoid the per-call map-lookup tax that would otherwise
fire on every Get / Set / Has.

Structural mutations (Spawn, Despawn, Attach, Detach) are queued
on the world's deferred queue and applied at known scheduler sync
points by `ApplyDeferred`. Component-value writes within an
entity's existing archetype are synchronous (they do not change
archetype membership).

### `engine/ecs/` root

Holds the `World` type and the typed call-site surface.

The `World` struct owns:
- An `*entity.Allocator`.
- The archetype table.
- The `locations` slice.
- The deferred queue.

The typed API surface — Spawn, Despawn, Attach, Detach, Get, Set,
Has, NewQuery — operates against `*World` and is the engine's
sole entry point into ECS. There is no parallel type-erased API.

## Entity allocator design

Settled material absorbed from the prior
`concepts/engine/entity-allocator.md`.

### Recycling

Forward-looking: each `Despawn` (after its deferred apply) returns
the entity's index to the allocator's recycle pool and increments
the generation table for that index. The motivation: without
recycling, any storage layout indexed by `entity.ID` would grow
unboundedly with cumulative spawns. With recycling, memory cost
stabilizes at peak-concurrent-entity-count rather than
cumulative-spawn-count. The allocator-level mechanism is in
`engine/ecs/entity/allocator.go`; the Despawn-side wiring lands
with the world API in a later session.

### Generations

Recycling creates an ABA hazard: code holding a reference to
"entity 42" from before despawn would silently alias a different
entity once 42 is recycled. Generations are the standard fix —
each (index, generation) pair is unique across the program's
lifetime (until the generation counter wraps at 2^32, which is
effectively never at voxel-game scale).

Forward-looking: every world API call taking an `entity.ID` will
validate via `Allocator.Validate`; a `false` result translates to
`ErrStaleEntity` at the world boundary. The cost is one index +
comparison per call.

### ApplyDeferred ordering

`ApplyDeferred` processes the queued structural mutations in a
single pass per frame, in queue order:

- Despawn supersedes any pending Attach/Detach for the same
  entity.
- Spawn + N pending Attaches against the same entity collapse to
  a single archetype placement (the apply logic walks the queue,
  composes the entity's final signature from spawn + attaches,
  and places once).
- Generation increment on Recycle happens at apply, not at the
  Despawn call site. The window between `Despawn(eid)` and the
  next `ApplyDeferred` is one where the entity remains readable
  and queryable; stale-handle errors fire only after the recycle
  completes.

## Storage strategy

Archetype is the engine's sole world backend. The typed surface
goes directly to archetype with no internal type-erased layer. The
per-call map-lookup tax that the cross-backend measurements
identified as the typed surface's open question is fixed at the
storage layer via the flat-array column index; the API surface
does not pay the boundary cost.

## Typed API surface

The surface comprises:

- `Spawn(w) entity.ID` / `Despawn(w, eid) error` — entity
  lifecycle, deferred.
- `Attach[T](w, eid, value) error` / `Detach[T](w, eid) error` —
  component lifecycle, deferred; upsert / no-op-if-absent.
- `Get[T](w, eid) (*T, error)` / `Set[T](w, eid, value) error` /
  `Has[T](w, eid) bool` — per-entity component access, synchronous.
- `NewQuery[V](w) *Query[V]` with `.All` range-over-func iteration
  — struct-based view query.

Sentinel errors per `code/conventions.md` §8: `ErrStaleEntity`,
`ErrNoComponent`.

## Deferred-mutation discipline

Structural mutations are queued and applied at known scheduler
sync points. The user-facing consequences:

- **Pointer lifetime.** Pointers returned by `Get` and by Query
  callbacks are valid until the next `ApplyDeferred`. Holding a
  pointer across an apply boundary risks accessing stale memory
  (the entity may have moved to a different archetype). Single-
  threaded execution makes apply boundaries deterministic; the
  rule is mechanical to follow at the human level. Go has no
  lifetime system to enforce it — this is documentation
  discipline.

- **Queued-despawn semantics.** A despawn that has been queued
  but not yet applied does not invalidate reads, queries, or
  subsequent Attach/Detach against the entity. Between
  `Despawn(eid)` and the next `ApplyDeferred`, the entity remains
  readable and queryable. Stale-handle errors fire only after
  recycle, which only happens at apply.

- **ApplyDeferred placement.** `ApplyDeferred` is a runtime
  concern called by the engine at known sync points in the
  frame; user-tier code does not call it directly. The exact
  placement of those sync points within the frame schedule is
  concept-tier in `concepts/engine/scheduler.md`.

## Forward-looking

- **Read/write encoding in Query views.** The struct-based view
  as designed encodes the *signature* (which components the query
  touches) but not the *direction* (read vs. write). The task
  declaration surface (`concepts/engine/scheduler.md` — Task
  declaration surface) requires read/write at the API for
  parallel-scheduling proof; the Query view is the call-site
  surface that must carry that encoding. This is the most pressing
  open question for the development session that lands Query.
  Approaches under consideration: marker types
  (`ReadOnly[Position]`), struct tags (`` `ecs:"read"` ``), naming
  convention. To be settled when the source code for Query is
  written.

- **Compaction policy for the entity allocator.** Recycling
  stabilizes memory at peak-concurrent-entity-count, but when
  the reserve queue grows large relative to live entities and
  storage arrays indexed by EntityID have correspondingly grown,
  reclaiming that memory needs an explicit compaction operation
  that reassigns live entity IDs to the low end of the ID space.
  Timeouts (expiring queued IDs) reintroduce monotonic ID growth
  and couple allocation to wall-clock time (nondeterministic and
  bad for reproducibility); compaction is structural — triggered
  by an explicit policy (e.g., max ID > 4× live count) at a
  frame boundary safe for ID reassignment. Trigger threshold,
  frequency, and the precise mechanism are open.

- **`engine/ecs/system/` sub-package.** Held per iterative depth.
  Its primitives are concept-tier in `concepts/engine/scheduler.md`;
  the package lands when those primitives firm up (and when the
  task declaration surface receives its source-code home).

- **Per-system custom storage.** Sparsesetgroup-style storage is
  preserved as a per-system option contingent on future evidence —
  but if a future system surfaces a measured need for it, it lives
  as a *separate specialized storage* outside the main world API
  (the Bevy-resources pattern: main archetype world plus per-system
  specialized storage as distinct APIs), not as a swap-in backend
  behind the existing surface. The trigger condition is a system
  with measured iteration-pattern need that the main archetype
  world doesn't serve well.

- **Archetype-table memory growth past 100k entities.** Carried
  forward from the storage-strategy commitment. Per-archetype page
  sizing and reclaim-on-shrink are unmeasured.

- **Threading evolution.** Single-threaded semantics are the
  current commitment; the path to within-frame parallelism via the
  task declaration surface's read/write declarations is
  concept-tier in `concepts/engine/scheduler.md`. Render-thread
  separation is concept-tier in `concepts/engine/render-thread.md`.

- **Bundle-struct Spawn convenience layer.** Deferred from the
  initial API design. Add as an additive feature on top of the
  basic API if a measured ergonomic need surfaces.

- **`MustGet[T]` panicking variant.** Deferred from the initial
  API design. Add if friction surfaces.

- **Deserialization path for `entity.ID`.** The only legitimate
  source of a valid `entity.ID` today is `Allocator.Allocate()`.
  Save/load support (via `engine/storage/`) will need a path that
  hands a deserialized ID back through the allocator. To be
  designed when persistence lands.
