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
   Structural mutation is dense-array push/pop.

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

## Finding

(Empty — the experiment has not yet run.)
