# ECS Storage Comparison

## Question

Which ECS storage strategy should the engine adopt first, and on what
evidence?

Three production-grade approaches dominate the space — archetype-based
(Bevy, flecs primary mode, Unity DOTS), sparse-set (EnTT), and
sparse-set with opt-in groups — and they trade off cleanly along
well-understood axes: iteration throughput, structural-mutation cost,
multi-component query cost, and memory behavior at scale. Which profile
fits a voxel-engine-shaped workload — moderate-to-high entity counts,
rich entities with broad component sets, significant structural churn
from projectiles, particles, and chunk-driven spawn/despawn — is not
settleable on paper. The engine's storage choice is load-bearing (the
layout ripples through physics and rendering per
`design/engine/runtime.md` — Inner-tier members, ECS paragraph), so the
decision needs measurement.

This experiment produces those measurements. It serves the open
question in `concepts/engine/ecs-storage.md`.

## Approach

### Shared interface

All three backends implement the same minimal interface so the
comparison measures storage-layer behavior rather than interface-design
differences. The interface bakes in the access patterns the real engine
will demand: read/write distinction at declaration time and a
direct-vs-deferred split for data mutation vs. structural mutation.

```
Spawn(components ...Component) EntityID    // immediate
Despawn(id EntityID)                       // deferred
Attach(id EntityID, c Component)           // deferred
Detach[T Component](id EntityID)           // deferred
Read[T Component](id EntityID) (T, bool)   // immediate
Write[T Component](id EntityID, value T)   // immediate
Query(componentSet) Iterator               // immediate, call-scoped
ApplyDeferred()                            // applies queued structural changes
```

`ApplyDeferred` is called between workload stages.

### Backends

1. **Archetype.** Entities grouped by exact component set; per-archetype
   dense arrays per component type. Iteration walks every archetype
   whose component set is a superset of the query. Structural mutation
   moves the entity to a different archetype.

2. **Sparse-set (no groups).** Per-component-type dense array plus
   sparse mapping from entity ID to dense index. Single-component
   iteration walks one dense array; multi-component iteration picks a
   driving set and indirects through sparse maps for the others.
   Structural mutation is dense-array push/pop. Two variants were
   measured at iteration-baseline fidelity to isolate the contribution
   of the sparse-mapping representation:
   - **Map variant** (`sparsesetmap/`) — `map[EntityID]int` for the
     sparse mapping. Standard Go primitive; supports any EntityID
     distribution; per-probe cost includes the map's hash function.
   - **Slice variant** (`sparsesetslice/`) — `[]int32` indexed by
     EntityID, sentinel `-1` for absence. Faster per-probe (direct
     index, no hash); memory cost is O(max EntityID) rather than
     O(entity count), so it relies on the entity allocator
     (`concepts/engine/entity-allocator.md`) to keep IDs dense.

   The slice variant carries forward as the active sparse-set backend
   for remaining workloads. The map variant is preserved as the
   implementation behind the iteration-baseline measurements (see
   *Finding*, 2026-05-06) but is not built out further — production
   sparse-set engines (EnTT and similar) use paged sparse arrays,
   not hash maps, so its further build-out would carry implementation
   cost without informational value beyond what the iteration data
   already captured.

3. **Sparse-set with opt-in groups.** Same as (2) but with explicit
   groups: declared component-set queries that keep their dense arrays
   sorted in lockstep, giving archetype-style iteration locally for the
   declared queries at the cost of restoring some structural-mutation
   overhead within the group.

Each backend implements *only enough* to support the workloads.
Production-grade generality would defeat the experiment's purpose.

### Workloads

1. **Iteration baseline.** N entities, each carrying
   `(Position, Velocity)`; integrate motion every frame. Measures raw
   single-archetype iteration throughput.

2. **Multi-component query.** N entities with varied component sets;
   query `(Position, Velocity, Health)` over the subset that has all
   three. Measures multi-component iteration cost — the axis on which
   archetype and sparse-set diverge most sharply.

3. **Structural churn.** Spawn and despawn entities per frame at
   varying rates. Measures deferred-command-buffer overhead and
   archetype-move cost.

4. **Attach/detach churn.** Existing entities gain or lose components
   per frame. Measures archetype-move cost specifically: sparse-set
   is nearly free on this axis, archetype is not.

5. **Mixed workload.** Combined frame — some integration, some
   queries, some spawning, some attach/detach — approximating a slice
   of a real voxel-game frame.

### Scales

Run each workload at 1k, 10k, and 100k entities. Higher scales (1M+)
added if performance permits and the lower scales surface meaningful
divergence.

### Metrics

Per (workload × scale × backend):

- Wall-clock time per frame (median, p50/p95/p99 across many frames)
- Allocations per frame (`runtime.MemStats`)
- Peak working-set memory

Output: one CSV per workload-scale-backend combination, plus a brief
markdown summary in this directory once analysis runs.

### Out of scope

- Threading, parallel scheduling, and concurrent access patterns.
  Single-threaded comparison.
- GPU or rendering-adjacent measurements. Storage layer only.
- Code quality, idiomaticity, conventions. Experiments are exempt from
  `code/conventions.md` per D-012.
- Backend generality beyond what the workloads require.

### Reactivity

If friction surfaces emerge that affect all three approaches equally —
implying the real bottleneck is elsewhere — or if a novel hybrid
surfaces naturally from implementation, the analysis writeup notes
those findings explicitly. The experiment's value is the measurements,
but it can also surface unforeseen design pressure on the storage
question.

## Runtime

Single Go module at `experiments/ecs-storage-comparison/`. Package layout:

- `storage/` — type-erased `Storage` and `Iterator` interfaces, `Signature` primitive (uint64 bitmask), generic helpers (`Read[T]`, `Write[T]`, `Attach[T]`, `Detach[T]`, `Spawn1`/`Spawn2`/`Spawn3`, variadic `Spawn`, `ComponentValueFor[T]`).
- `archetype/` — first backend. Entities grouped by exact component set; per-archetype byte-slice columns; iteration walks signature-matching archetypes via the package-internal `iterator` type.
- `sparsesetmap/`, `sparsesetslice/` — backend 2 in two variants (map and slice sparse representations); see *Approach > Backends*.
- `sparsesetgroup/` — backend 3, to follow.
- `workload/` — workload definitions. Currently iteration baseline (`IterationSetup`, `IterationTick`); others land alongside as needed.
- `main.go` — flag-driven harness. Constructs the chosen backend, runs the workload's setup, ticks N frames while capturing per-frame timing, writes CSV plus a stdout summary.
- `results/` — CSV output directory (gitignored).

### How to run

From the experiment directory:

```
go run . -backend=archetype -workload=iteration -scale=1000 -frames=1000
```

All flags optional; defaults shown above. Valid values:

- `-backend` — `archetype | sparsesetmap | sparsesetslice` (`sparsesetgroup` to follow).
- `-workload` — `iteration` (others to follow).
- `-scale`, `-frames` — any positive integer.
- `-out` — output directory (default `results`).

Output: one CSV per `(backend, workload, scale)` combination at `{out}/{backend}_{workload}_{scale}.csv`. Stdout one-line summary with per-run allocation totals and peak heap.

### How to interpret results

CSV columns:

- `frame` — zero-indexed frame number.
- `time_ns` — wall-clock tick duration captured via `time.Now()`/`time.Since()`.

Distribution stats (median, p50/p95/p99) are computed offline — the harness leaves raw per-frame data for analysis flexibility.

Stdout summary fields:

- `frames` — frame count for the run.
- `allocs/frame` — average heap allocations from `runtime.MemStats.Mallocs` deltas across the run.
- `bytes/frame` — average heap bytes from `TotalAlloc` deltas.
- `peak_heap` — `HeapInuse` after the run completes (post-tick resident heap high-water).

**Cache residency caveat (iteration baseline workload).** Per entity, hot data is 24 bytes (Position + Velocity). Total hot data:

- 1k entities → ~24 KB → L1d-resident on typical desktops (32 KB L1d on Intel Coffee Lake or similar Zen).
- 10k entities → ~240 KB → fits L2 (256 KB to 1 MB on consumer CPUs) but exceeds L1.
- 100k entities → ~2.4 MB → exceeds typical L2; hits L3 (8 MB+) or main memory.

Cross-backend comparison is most informative at 10k+ scales where storage layout drives the cost. At 1k, all backends compress to similar numbers because the working set fits in L1. Allocation differences surface at any scale since they're independent of cache.

## Finding

### 2026-05-04 — first run, archetype × iteration baseline × 1000 entities × 1000 frames

Test machine: Intel i7-9700K @ 4.9 GHz, 32 KB L1d, 32 GB DDR4.

Distribution: min=8954 ns, p50=9030, p95=9491, p99=11637, max=16416. Per-entity cost ~9.0 ns (~44 cycles at 4.9 GHz). Distribution very tight — p99 only 28% above p50.

Allocations: 3.00/frame, 68.19 bytes/frame. Decomposes exactly to the slice literal `[]ComponentID{posID, velID}` in `IterationTick` (1 alloc), the `matches` slice in `archetype.Storage.Query` (1 alloc), and `&iterator{...}` returned from Query (1 alloc). All inherent to the current Query API; reducing them would require pooling or a builder pattern. Constant across all three backends, so it shouldn't bias the comparison.

Peak heap: ~632 KB — dominated by setup (1000 entities × Position + Velocity + archetype overhead).

**Caveat — L1-resident at this scale.** Hot data is ~24 KB (well within the 32 KB L1d cache), so this measurement reflects compute + L1 throughput rather than storage layout. Cross-backend comparison at 1k will be uninformative; the meaningful comparison happens at 10k+ where the working set is L2/L3-resident.

**Status.** Archetype implementation correct, harness producing sane numbers. Sparse-set and sparse-set-with-groups backends pending. Multi-component query, structural churn, attach/detach churn, and mixed workloads pending.

### 2026-05-06 — sparsesetmap + sparsesetslice landed; iteration baseline three-way at 1k/10k/100k

Test machine: same as prior entry (Intel i7-9700K @ 4.9 GHz, 32 KB L1d, 32 GB DDR4).

Per-entity cost (p50 ÷ scale, ns/entity):

| Scale | archetype | sparsesetmap | sparsesetslice |
|-------|-----------|--------------|----------------|
| 1k    | 9.03      | 18.87        | 8.94           |
| 10k   | 8.93      | 23.95        | 8.81           |
| 100k  | 8.73      | 31.35        | 8.67           |

Distribution at 100k (ns):

| Backend         | min     | p50     | p95     | p99     | max     |
|-----------------|---------|---------|---------|---------|---------|
| archetype       |  869764 |  873192 |  919362 | 1087472 | 1125277 |
| sparsesetmap    | 3034075 | 3135222 | 3283751 | 3408938 | 4127463 |
| sparsesetslice  |  859645 |  866754 |  914936 | 1000348 | 1044484 |

**Headline.** Sparse representation choice carried the entire iteration gap. sparsesetslice tracks archetype within ~1% across all measured scales; sparsesetmap shows the cache-cliff growth (8.94 → 18.87 → 23.95 → 31.35 ns/entity) that signals hash-probe randomization defeating the prefetcher.

**Why.** The iteration baseline spawns entities 1..N in deterministic order. Driving on Position and walking its `entities` array yields IDs 1, 2, 3, ..., N in sequence. For sparsesetslice, the non-driver probe `Velocity.sparse[entity]` becomes a sequential read into a `[]int32` at positions 1..N — three concurrent sequential streams (driver entities, non-driver sparse, non-driver dense) the prefetcher handles trivially. For sparsesetmap, the hash function within the map's bucket table randomizes the access pattern, breaking prefetch and producing the per-entity cost growth visible at 10k+ where the working set spills out of L1.

**Allocation parity held.** All three backends at 3.00 allocs/frame — the slice literal in workload's `IterationTick`, the per-Query slice (`matches []*archetype` for archetype, `refs []queryRef` for both sparse-set variants), and `&iterator{}`. Bytes/frame divergence (archetype 68.19, both sparse variants 92.19) reflects the queryRef slice carrying twice the per-element width of archetype's `*archetype` matches slice, populated with two refs vs one match.

**Heap profile at 100k.** archetype 7.91 MB, sparsesetmap 10.0 MB, sparsesetslice 5.11 MB. archetype's ~4.8 MB lives in its `locations` map (entity → archetype + row, populated by Spawn) and dominates its overhead; sparse-set has no equivalent at this implementation fidelity because each column tracks its own entities. *Caveat:* a full sparse-set implementation supporting Despawn/Attach/Detach may need similar per-entity tracking, so the heap delta will partially close once those methods land across the backends.

**Implication for the engine decision.** Iteration is no longer the deciding axis between archetype and sparse-set — at least for this workload class with dense ID distributions. The decision shifts to: (a) whether voxel-game ID distributions stay dense, which depends on the entity allocator's recycling behavior (`concepts/engine/entity-allocator.md`); (b) the remaining four workloads — especially attach/detach churn, where sparse-set is structurally expected to win because each column is independent and archetype must move entire rows between archetypes; (c) whether sparsesetslice's iteration parity holds for general queries (multi-component varied workload still pending — the bounds check in `iterator.Next` never fired in the iteration baseline because every entity carried both queried components).

**Status.** All three current backends at iteration-baseline fidelity (Spawn + Query + iterator filled; Despawn / Attach / Detach / Read / Write / ApplyDeferred stubbed). Cross-backend iteration data captured at 1k/10k/100k. **sparsesetmap is culled from further build-out** — its instrumentation purpose (isolating the sparse-mapping representation's contribution to iteration cost) is complete; it does not represent a production sparse-set design (real engines use paged sparse arrays, not hash maps) and the carrying cost of filling stubs and running additional workloads would not buy informational value beyond what today's data captured. Subsequent work targets archetype + sparsesetslice + sparsesetgroup across the remaining four workloads (multi-component query, structural churn, attach/detach churn, mixed) and the remaining six interface methods. Next session: sparsesetgroup at iteration-baseline fidelity.
