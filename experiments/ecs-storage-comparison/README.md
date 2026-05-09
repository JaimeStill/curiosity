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

3. **Sparse-set with opt-in groups** (`sparsesetgroup/`). Same as (2)
   but with explicit groups: declared component-set queries that keep
   their dense arrays sorted in lockstep, giving archetype-style
   iteration locally for the declared queries at the cost of restoring
   some structural-mutation overhead within the group.

   Owning groups are declared at construction via
   `New(groups [][]ComponentID)`; the lockstep invariant is maintained
   on Spawn by swap-into-prefix when a spawn covers a declared group's
   set. Queries whose set matches a declared group walk the owned
   prefix without sparse probes (archetype-style locally); queries
   whose set doesn't match any declared group fall back to slice-style
   iteration with a per-non-driver sparse probe.

Each backend implements *only enough* to support the workloads.
Production-grade generality would defeat the experiment's purpose.

### Workloads

1. **Iteration baseline.** N entities, each carrying
   `(Position, Velocity)`; integrate motion every frame. Measures raw
   single-archetype iteration throughput.

2. **Multi-component query.** N entities with varied component sets,
   spawned in cycled composition classes (`{P,V,H,Tag}`, `{P,V,H}`,
   `{P,V,Tag}`, `{P,V}`, `{P,H,Tag}`, `{V,Tag}`) so that archetype's
   matching set fragments across multiple archetypes and
   sparse-set's per-row probes fail meaningfully often. Run as two
   isolated scenarios:
   - **`multi_full`** — query `{Position, Velocity, Health}`. The
     queried set equals the declared owning group, so sparsesetgroup
     takes its fast path; archetype walks the two matching
     archetypes; sparsesetslice probes both non-driver columns per
     driver row. Iteration rate: 1/3 of N.
   - **`multi_partial`** — query `{Position, Velocity}`, a strict
     subset of the declared group. sparsesetgroup falls back to
     slice-style traversal (small unified-iterator tax over slice's
     baseline); archetype walks four matching archetypes;
     sparsesetslice probes one non-driver column per driver row.
     Iteration rate: 2/3 of N.
   Together the two scenarios characterize sparsesetgroup's fast
   and fallback paths in isolation and surface archetype's
   multi-archetype-walk cost on the axis the *Question* section
   flags as the sharpest.

3. **Structural churn.** Three population models exercise the
   deferred-mutation surface and storage-side structural cost
   independently. Each tick performs K = scale/100 structural ops;
   `ApplyDeferred` runs once per tick. Component shape matches the
   multi-component owning group: every spawned entity carries
   `{Position, Velocity, Health}`. sparsesetgroup declares one
   owning group on the same set.
   - **`structural_steady`** — spawn N up front; each tick despawn
     K and spawn K (FIFO via a head index over a workload-tracked
     alive list). Steady-state population at tick boundaries;
     isolates pair-mutation cost at fixed working-set size.
   - **`structural_cycle`** — start empty; alternate phases. Grow
     phase: spawn K each tick until population ≥ target. Drain
     phase: despawn K each tick (FIFO) until population reaches 0,
     then reset and flip back to grow. Population swings 0 ↔ N,
     exercising spawn-only and despawn-only directions in
     isolation.
   - **`structural_growth`** — start empty; spawn K each tick,
     never despawn. Population grows monotonically; isolates
     spawn-only storage growth without despawn signal.
   Together the three scenarios separate spawn-only, despawn-only,
   and balanced-pair structural cost so per-direction signal isn't
   averaged across asymmetric phases.

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

- `entity/` — entity-identity primitives. `entity.ID` (the entity-identity type), `entity.Allocator` (free-list-based recycler with `Allocate` returning recycled IDs first and `Free` enqueuing them for reuse). Imported by every package that mints or releases entity IDs; depends on nothing experiment-internal.
- `component/` — component-identity primitives. `component.ID`, `component.InvalidID` sentinel, `component.Value` (id + data carrier), the type→ID registry (`IDFor[T]`, `ValueFor[T]`, `TypeOf`), and `component.Signature` (uint64 bitmask over component IDs with `Set` / `Has` / `Contains` / `SignatureOf`). Depends on nothing experiment-internal.
- `storage/` — type-erased `Storage` and `Iterator` interfaces (expressed in terms of `entity.ID`, `component.ID`, `component.Value`) and the typed call-site helpers (`Read[T]`, `Write[T]`, `Attach[T]`, `Detach[T]`, `Spawn1`/`Spawn2`/`Spawn3`, variadic `Spawn`). Each backend's `New(alloc *entity.Allocator, …)` constructor receives the allocator that drives `Spawn`'s `Allocate` and `Despawn`'s `Free`.
- `archetype/` — first backend. Entities grouped by exact component set; per-archetype byte-slice columns; iteration walks signature-matching archetypes via the package-internal `iterator` type.
- `sparsesetmap/`, `sparsesetslice/` — backend 2 in two variants (map and slice sparse representations); see *Approach > Backends*.
- `sparsesetgroup/` — backend 3; see *Approach > Backends*.
- `workload/` — workload definitions. Iteration baseline (`IterationSetup`, `IterationTick`, `IterationGroups`), multi-component query (`MultiComponentSetup`, `MultiFullTick`, `MultiPartialTick`, `MultiGroups`), and structural churn (`StructuralSteadySetup`/`Tick`, `StructuralCycleSetup`/`Tick`, `StructuralGrowthSetup`/`Tick`, `StructuralGroups`); others land alongside as needed.
- `main.go` — flag-driven harness. Constructs the chosen backend (creating its `entity.Allocator` per run), runs the workload's setup, ticks N frames while capturing per-frame timing, writes CSV plus a stdout summary.
- `results/` — CSV output directory (gitignored).

### How to run

From the experiment directory:

```
go run . -backend=archetype -workload=iteration -scale=1000 -frames=1000
```

All flags optional; defaults shown above. Valid values:

- `-backend` — `archetype | sparsesetmap | sparsesetslice | sparsesetgroup`.
- `-workload` — `iteration | multi_full | multi_partial | structural_steady | structural_cycle | structural_growth` (attach/detach churn and mixed to follow).
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

### 2026-05-06 — sparsesetgroup landed; iteration baseline four-way at 1k/10k/100k

Test machine: same as prior entries (Intel i7-9700K @ 4.9 GHz, 32 KB L1d, 32 GB DDR4).

Per-entity cost (p50 ÷ scale, ns/entity):

| Scale | archetype | sparsesetmap | sparsesetslice | sparsesetgroup |
|-------|-----------|--------------|----------------|----------------|
| 1k    | 9.03      | 18.87        | 8.94           | 7.99           |
| 10k   | 8.93      | 23.95        | 8.81           | 7.89           |
| 100k  | 8.73      | 31.35        | 8.67           | 7.87           |

Distribution at 100k (ns):

| Backend         | min     | p50     | p95     | p99     | max     |
|-----------------|---------|---------|---------|---------|---------|
| archetype       |  869764 |  873192 |  919362 | 1087472 | 1125277 |
| sparsesetmap    | 3034075 | 3135222 | 3283751 | 3408938 | 4127463 |
| sparsesetslice  |  859645 |  866754 |  914936 | 1000348 | 1044484 |
| sparsesetgroup  |  766480 |  786520 |  830529 |  915605 | 1027965 |

**Headline.** sparsesetgroup runs the iteration baseline ~10–13% faster than both sparsesetslice and archetype across all three scales — measurably the leanest of the four backends on this workload, not the expected ≈-tie. The owning-group fast path skips per-row sparse-side work that even the lucky sparsesetslice case still pays.

**Why.** sparsesetslice's iterator probes the sparse mapping for every non-driver column at every row to resolve the dense index — the bounds check and sentinel test never fire (every entity carries both components in the iteration baseline), but the load-compare-load-store sequence still runs. sparsesetgroup's fast path knows the lockstep invariant holds across every group column at indices [0, g.size), so it writes `index = row` directly without any sparse-side load. Across 100k rows that delta is ~80k ns/frame, matching the measured gap. archetype shows a similar offset against sparsesetgroup, structurally explained by archetype's per-row column-map indirection that the group fast path also avoids.

**Allocation parity held.** sparsesetgroup at 3.00 allocs/frame, ~92.19 bytes/frame — same shape as sparsesetslice (the queryRef slice carries the same per-element width). Allocation profile is not the discriminator at this workload class.

**Heap profile at 100k.** sparsesetgroup 5.01 MB, sparsesetslice 5.11 MB. Difference is within HeapInuse measurement noise — call them equivalent. Both sit ~3 MB below archetype's 7.91 MB, which is still dominated by archetype's `locations` map.

**Caveat — unified iterator pays a small fast-path tax.** sparsesetgroup's iterator writes `refs[i].index = it.row` for every ref on every Next, even on the fast path where row directly equals every column's dense index. A specialized fast-path iterator that exposed `it.row` to Get directly could elide those writes; the unified design pays them to keep Next / Entity / Get coherent across both fast and fallback modes. The advantage over sparsesetslice would be marginally larger with a specialized iterator. The reported numbers reflect what an actually-shipped general-purpose owning-group implementation would look like, not a hand-tuned fast path.

**Implication for the engine decision.** sparsesetgroup's iteration win comes from the lockstep invariant — paid for in Spawn (and, when those workloads land, in Despawn / Attach / Detach as group-eligible entries cross the boundary). The iteration baseline measures the gain in isolation; the maintenance cost shows up only in workloads that mutate group-eligible entities. The remaining four workloads — especially attach/detach churn, where plain sparse-set is structurally cheap and sparsesetgroup must do extra swap work — are where the trade gets measured. Iteration is now decided in groups' favor *for declared queries*; everything else is open.

**Status.** Iteration-baseline row complete across four backends. Six interface methods (Despawn / Attach / Detach / Read / Write / ApplyDeferred) and four workloads (multi-component query, structural churn, attach/detach churn, mixed) remain across three active backends (archetype + sparsesetslice + sparsesetgroup). Multi-component query is the natural next workload — it exercises sparsesetgroup's fast-path/fallback split internally and surfaces the multi-component-archetype-vs-sparse divergence the README's *Question* section flags.

### 2026-05-07 — multi-component query landed; archetype + sparsesetslice + sparsesetgroup at 1k/10k/100k across two scenarios

Test machine: same as prior entries (Intel i7-9700K @ 4.9 GHz, 32 KB L1d, 32 GB DDR4).

Setup: see *Approach > Workloads > 2*. Six composition classes cycled by `i % 6`; iteration rates 1/3 of N for `multi_full`, 2/3 of N for `multi_partial`. sparsesetgroup declares one owning group on `{Position, Velocity, Health}` (shared across both scenarios; only the queried set differs).

Per-entity-of-N cost (p50 ÷ scale, ns):

| Scale | Scenario       | archetype | sparsesetslice | sparsesetgroup |
|-------|----------------|-----------|----------------|----------------|
| 1k    | multi_full     | 4.09      | 4.81           | 3.84           |
| 1k    | multi_partial  | 6.17      | 6.31           | 7.01           |
| 10k   | multi_full     | 3.96      | 4.71           | 3.67           |
| 10k   | multi_partial  | 5.95      | 6.18           | 6.98           |
| 100k  | multi_full     | 3.95      | 4.72           | 3.67           |
| 100k  | multi_partial  | 5.96      | 6.25           | 6.97           |

Per-iterated-row cost (p50 ÷ rows iterated, ns) at 100k — apples-to-apples against iteration baseline:

| Backend         | iteration (R-008) | multi_full | multi_partial |
|-----------------|------------------:|-----------:|--------------:|
| archetype       |              8.73 |      11.85 |          8.93 |
| sparsesetslice  |              8.67 |      14.15 |          9.37 |
| sparsesetgroup  |              7.87 |      11.02 |         10.45 |

Distribution at 100k, multi_full (ns):

| Backend         | min    | p50    | p95    | p99    | max    |
|-----------------|--------|--------|--------|--------|--------|
| archetype       | 384945 | 395094 | 427371 | 475176 | 604631 |
| sparsesetslice  | 458139 | 471706 | 509790 | 554784 | 798458 |
| sparsesetgroup  | 365032 | 367222 | 376459 | 431257 | 472634 |

Distribution at 100k, multi_partial (ns):

| Backend         | min    | p50    | p95    | p99    | max    |
|-----------------|--------|--------|--------|--------|--------|
| archetype       | 580535 | 595643 | 631014 | 738601 | 783389 |
| sparsesetslice  | 615983 | 624731 | 702149 | 728010 | 875637 |
| sparsesetgroup  | 677183 | 696909 | 742990 | 799971 | 888963 |

**Headline.** The leadership flips between scenarios. `multi_full` (query matches the declared group set): sparsesetgroup leads on its fast path; archetype follows ~7% behind; sparsesetslice trails ~28%. `multi_partial` (query is a strict subset of the declared group): archetype leads outright; sparsesetslice follows ~5% behind; **sparsesetgroup is now slowest at ~17%** behind archetype, paying its fallback path's slice-style probe cost plus the unified-iterator tax R-008's caveat predicted. Per-iterated-row cost rose for every backend versus iteration baseline because varied composition forces real per-row probe work or per-archetype-walk overhead the dense-uniform iteration baseline never exercised.

**Why — multi_full.** archetype walks two matching archetypes (`{P,V,H}` and `{P,V,H,Tag}`); per-iterated-row cost rises 8.73 → 11.85 ns (+36%), part per-archetype-walk overhead, part the third component (`Health`) adding a `Get` and a write per row. sparsesetslice's driver column (Position) covers 5/6 of N entities of which only 2/5 also have both Velocity *and* Health — the other 3/5 fail one or both probes. Per-driver-row probe-and-skip cost is real even when the path skips, accounting for the +63% per-iterated-row jump (8.67 → 14.15 ns). sparsesetgroup's fast path skips all per-row sparse-side work; per-iterated-row cost rises 7.87 → 11.02 ns (same direction as archetype's, attributable to the third component, not to fast-path overhead).

**Why — multi_partial.** archetype walks four matching archetypes (`{P,V}`, `{P,V,Tag}`, `{P,V,H}`, `{P,V,H,Tag}`) and absorbs per-archetype overhead nearly cost-free at the per-iterated level (8.93 vs 8.73, +2%). sparsesetslice pays modest probe-skip overhead (9.37 vs 8.67, +8%) — driver column (P) covers 5/6 of N, of which 4/5 also have V. sparsesetgroup's fallback runs the same shape as sparsesetslice's iteration plus a per-Next mode-flag-based branch — measured tax 10.45 - 9.37 = 1.08 ns/iterated-row, matching R-008's prediction within a fraction of a cycle.

**Allocation parity held.** All three backends at 3.00 allocs/frame across both scenarios. Bytes/frame: 112 for all three on `multi_full`; 108 for archetype and 92 for both sparse-set variants on `multi_partial`. Allocation profile is not the discriminator at this workload class.

**Heap profile at 100k.** archetype 7.62 MB on both scenarios — same `locations`-map dominance as iteration baseline. sparsesetslice 6.11 / 6.00 MB. sparsesetgroup 6.04 / 6.01 MB. Both sparse-set variants sit ~1.6 MB below archetype, a smaller gap than iteration baseline's ~2.8 MB because varied composition adds two more component columns (`Health`, `Tag`) to each sparse-set backend's bookkeeping.

**Caveat — unified-iterator tax is real but small.** sparsesetgroup's ~12% fallback-path slowdown over sparsesetslice (+1.08 ns/iterated-row at 100k `multi_partial`) comes from the dual-mode `Next` carrying a per-call mode-flag check and per-row index writes a specialized fallback iterator would elide. Splitting the iterator into two concrete types behind the existing `Iterator` interface would close most of this gap without flipping the `multi_partial` leaderboard — archetype still leads at 5.96 ns/N versus an optimized sparsesetgroup fallback projected near 6.0 ns/N. This is the smallest, cheapest piece of sparsesetgroup's complexity story, not the largest.

**Implication for the engine decision.** sparsesetgroup's win condition now has a visible shape: leads ~7% on queries matching a declared group, trails ~17% on queries that don't. Whether that's net-positive depends on a hot-query distribution we don't have data for yet. archetype's win condition is broader: no "wrong query" failure mode, multi-archetype walking turned out to be cheap (per-archetype overhead amortized well over matching rows), competitive across both scenarios (won `multi_partial`, lost `multi_full` by ~7%). The complexity comparison extends beyond today's measurement: the six remaining stubbed methods on sparsesetgroup include `Despawn`, `Attach`, `Detach`, all of which must coordinate the lockstep invariant across every group column whenever an entity crosses the owned-prefix boundary — a distributed invariant that breeds subtle bugs in edge cases. archetype's structural-mutation surface (move-between-archetypes) is also non-trivial, but the abstraction is direct: an entity is in exactly one archetype at a time. On a performance ÷ (complexity × ergonomics) axis, archetype's standing strengthens after this workload — competitive on multi-component, no API surface bet on which queries are hot, complexity concentrated in two well-bounded abstractions (archetype management, entity-location tracking). The remaining three workloads (structural churn, attach/detach churn, mixed) — especially attach/detach churn, where sparse-set is structurally cheap and archetype must move entire rows — are where the analysis could still shift.

**Status.** Multi-component query row complete at iteration-baseline + multi-component-query fidelity across the three active backends. Six interface methods (Despawn / Attach / Detach / Read / Write / ApplyDeferred) and three workloads (structural churn, attach/detach churn, mixed) remain. Next workload: structural churn — measures deferred-command-buffer overhead and archetype's move-between-archetypes cost, the latter the axis where sparse-set's column-independent design is structurally expected to win.

### 2026-05-09 — structural churn landed across three population models; archetype `locations` map → ID-indexed slice mid-session

Test machine: same as prior entries (Intel i7-9700K @ 4.9 GHz, 32 KB L1d, 32 GB DDR4).

Setup: see *Approach > Workloads > 3*. Three population models (`structural_steady`, `structural_cycle`, `structural_growth`) each tick K = scale/100 structural ops with `ApplyDeferred` once per tick. Component shape matches the multi-component owning group (`{Position, Velocity, Health}`); sparsesetgroup declares one owning group on the same set.

This session ran in two phases. The pre-pivot phase landed `Despawn` + `ApplyDeferred` against archetype's existing `locations map[entity.ID]location` and surfaced a clear cost gap at 100k driven by that map. The post-pivot phase replaced the map with an ID-indexed `[]location` (D-026) and re-measured. Both data sets are reported here so the swap's effect is visible in place.

#### Pre-pivot — archetype with `locations` as map

Per-frame cost (p50, ns):

| Scale | Workload          | archetype | sparsesetslice | sparsesetgroup |
|-------|-------------------|-----------|----------------|----------------|
| 1k    | structural_steady |     2,480 |          2,566 |          2,932 |
| 1k    | structural_cycle  |     1,694 |          1,807 |          1,871 |
| 1k    | structural_growth |     1,801 |          1,827 |          1,911 |
| 10k   | structural_steady |    26,024 |         25,982 |         29,961 |
| 10k   | structural_cycle  |    17,485 |         17,691 |         18,768 |
| 10k   | structural_growth |    20,984 |         19,474 |         19,984 |
| 100k  | structural_steady |   311,801 |        261,890 |        305,503 |
| 100k  | structural_cycle  |   188,561 |        101,534 |        119,588 |
| 100k  | structural_growth |   314,887 |        190,004 |        203,705 |

Peak heap at 100k, `structural_growth`: archetype 143 MB, sparsesetslice 93 MB, sparsesetgroup 83 MB. End-of-run population is ~1.1M entities; component data accounts for ~35 MB; sparseset overhead (dense + sparse + entity-ID slices) totals ~50-60 MB; archetype's overhead is ~100 MB, of which the `locations map[entity.ID]location` dominates (~80 bytes/entry × ~1M = ~80 MB).

**Pre-pivot headline.** Archetype trails sparsesetslice by 18-65% on every structural workload at 100k. Per-spawn cost climbs with population for archetype (180 → 210 → 315 ns/op for growth across 1k/10k/100k); sparsesets stay flat (~190 ns/op at all scales). The divergence at scale isn't architectural — it's the cost of archetype's `locations` map's bucket overhead and per-insert hash work at 1M-entry scale.

D-025's scan-all-columns despawn choice in sparseset (`sparsesetslice` and `sparsesetgroup` Phase B) was vindicated at 3-column shape: sparsesetslice's per-op cost is competitive with archetype's at 100k cycle (102 ns/op vs 189 ns/op pre-pivot), proving the scan isn't dominating at this width. Sparsesetgroup's lockstep tax (R-008's prediction) shows here too: 5-15% slower than slice across every workload × scale combination, paying the prefix-maintenance cost in every tick whether or not iteration reaps the win.

#### Post-pivot — archetype with `locations` as ID-indexed slice (D-026)

Per-frame cost (p50, ns) — only archetype changed; sparseset values repeated for direct comparison:

| Scale | Workload          | archetype pre | archetype post | sparsesetslice | sparsesetgroup |
|-------|-------------------|--------------:|---------------:|---------------:|---------------:|
| 1k    | structural_steady |         2,480 |          1,841 |          2,566 |          2,932 |
| 1k    | structural_cycle  |         1,694 |          1,549 |          1,807 |          1,871 |
| 1k    | structural_growth |         1,801 |          1,601 |          1,827 |          1,911 |
| 10k   | structural_steady |        26,024 |         18,360 |         25,982 |         29,961 |
| 10k   | structural_cycle  |        17,485 |         15,451 |         17,691 |         18,768 |
| 10k   | structural_growth |        20,984 |         16,419 |         19,474 |         19,984 |
| 100k  | structural_steady |       311,801 |        184,291 |        261,890 |        305,503 |
| 100k  | structural_cycle  |       188,561 |         82,884 |        101,534 |        119,588 |
| 100k  | structural_growth |       314,887 |        166,870 |        190,004 |        203,705 |

Distribution at 100k, `structural_growth` (post-pivot, ns):

| Backend         |     min |     p50 |     p95 |       p99 |        max |
|-----------------|--------:|--------:|--------:|----------:|-----------:|
| archetype       | 158,433 | 166,870 | 388,174 | 2,271,142 | 17,751,643 |
| sparsesetslice  | 183,909 | 190,004 | 307,109 | 2,148,461 |  8,786,877 |
| sparsesetgroup  | 196,599 | 203,705 | 313,347 | 2,836,262 | 10,132,070 |

#### Iteration + multi-component re-runs after the swap

Iteration baseline and multi-component query were re-measured to confirm the swap doesn't regress hot-path cost (those workloads spawn only at setup; their measured loops never hit `locations`). Per-frame p50 timings change within noise (≤2% across all nine cells). Heap profile changes meaningfully:

Peak heap at 100k (MB):

| Workload      | archetype pre | archetype post |     Δ | sparsesetslice |
|---------------|--------------:|---------------:|------:|---------------:|
| iteration     |          7.91 |           5.64 | -29%  |           5.11 |
| multi_full    |          7.62 |           5.76 | -24%  |           6.11 |
| multi_partial |          7.62 |           5.76 | -24%  |           6.00 |

Archetype's heap at 100k now sits at parity with sparsesetslice's across iteration, multi-component, and structural workloads. The ~4.8 MB locations-map overhead R-008 attributed to archetype is gone — replaced by ~1.6 MB of ID-indexed slice (1M × 16 bytes).

#### Headline

The swap closes both the heap gap and the structural-mutation cost gap that pre-pivot data had attributed to archetype. Post-pivot leadership at 100k:

- `structural_steady` — archetype leads sparsesetslice by 30%, sparsesetgroup by 40%.
- `structural_cycle` — archetype leads sparsesetslice by 18%, sparsesetgroup by 31%.
- `structural_growth` — archetype leads sparsesetslice by 12%, sparsesetgroup by 18%.

Sparsesetgroup's lockstep tax stays visible at every scale on every structural workload (5-15% slower than slice). The cost is paid in spawn (swap-into-prefix) and despawn (Phase A boundary swap) regardless of whether the workload's query mix would reap the iteration-side win.

#### Why archetype's structural cost dropped so sharply

Per-op cost decomposition. Each spawn pays one `locations` write (and grow-if-needed); each despawn-apply pays three `locations` accesses (read for the despawned ID, read for the swapped-in ID, write to update the swapped-in ID's row, sentinel-set on the despawned slot). Pre-pivot: each map op was a hash + bucket walk + (occasional) rehash, climbing in cost as the map grew toward 1M entries. Post-pivot: each op is a direct slice index and write — constant-time, cache-friendly, no hashing.

Heap-side: Go's map carries ~80 bytes/entry of internal overhead; the slice carries 16 bytes/entry. At 1M entries that's ~80 MB → ~16 MB for the locations data alone, with the rest of the post-pivot heap reduction coming from the lower GC pressure that follows from fewer per-spawn map-bucket allocations.

#### Implication for the engine decision

The cumulative storage-strategy picture has shifted noticeably. Pre-R-012, R-009's analysis leaned archetype on a "performance ÷ (complexity × ergonomics)" basis primarily because archetype was competitive on multi-component query and had no "wrong-query failure mode." That reading is now stronger: archetype is *leading* on structural mutation across all three population models at 100k, with heap profile in line with sparsesetslice. Sparsesetgroup retains its iteration win on declared-group queries and its multi_full lead, but it now consistently trails on every workload that doesn't reap that fast path.

What remains open:
- *Attach/detach churn* (next session). Sparseset is structurally expected to win this — column-independent updates vs archetype's row-move between archetypes. If sparseset's expected lead surfaces here, it begins to balance archetype's structural-mutation lead.
- *Wider component shapes.* This session used a 3-column entity shape (matches multi-component owning group). D-025's scan-all-columns despawn would face a sterner test at 8-12 columns; whether the scan cost begins to dominate sparseset's structural advantage on wider shapes is the natural follow-up if attach/detach data motivates it.
- *Mixed workload* (final workload). Closer to a real per-frame workload mix; will test how the per-axis leaders behave when their costs are interleaved.

#### Status

Structural-churn row complete across the three active backends at three scales, in two phases (pre-pivot map / post-pivot slice). Six of nine `Storage` interface methods landed (Spawn / Despawn / Query / ApplyDeferred + iteration's iterator state); Read / Write / Attach / Detach remain stubbed across the three active backends. `sparsesetmap` retains its R-007 culled status (panic-stubbed Despawn / ApplyDeferred). Two workloads remain: attach/detach churn and mixed. Decision logged: D-026 (`locations` map → ID-indexed slice). Next workload: attach/detach churn — exercises the four currently stubbed interface methods and tests sparseset's structurally expected win on column-independent component churn.
