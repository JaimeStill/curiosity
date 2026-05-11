# workload

Workload definitions for the ECS storage comparison. Each workload
bundles three pieces: a `*Setup` function spawning N entities, one
or more `*Tick` functions performing the per-frame computation the
backends are measured against, and a `*Groups` function declaring
owning groups for the `sparsesetgroup` backend.

The harness in `main.go` selects a workload by name (`-workload=...`)
and dispatches setup, tick, and groups through `selectWorkload` and
`selectWorkloadGroups`. Backends other than `sparsesetgroup` ignore
the group declarations; the parameter exists in `newBackend` to keep
the dispatch uniform.

## Components

The package defines four component types, registered with the
`component` package's type registry on first reference
(`component.IDFor[T]()`). The same Position / Velocity /
Health / Tag IDs are reused across every workload.

- **`Position`** (`X, Y, Z float32`, 12 bytes) — simulation state.
  Mutated each tick by adding `Velocity`. The shape mirrors the
  smallest realistic per-entity payload an engine would carry.
- **`Velocity`** (`X, Y, Z float32`, 12 bytes) — simulation state.
  Read each tick.
- **`Health`** (`Current, Max int32`, 8 bytes) — third-component
  payload for multi-component workloads. The differing struct shape
  from Position/Velocity (two int32s vs three float32s) makes
  per-component reads visibly distinct in tooling and avoids
  accidental layout-aliasing across components.
- **`Tag`** (`V byte`, 1 byte) — marker carried by some entities to
  fragment archetype's matching set. `V` is unused payload; the
  presence of `Tag` in an entity's component set is the whole point.
  A single byte rather than an empty struct avoids any zero-size-type
  edge case in archetype's column-byte-slice math.

## iteration

The simplest workload. N entities each carrying `Position` and
`Velocity`. Per tick: query `{Position, Velocity}` and add velocity
to position.

Spawn pattern (every entity is the same composition):

```
i = 0:    Spawn2(Position{0, 0, 0}, Velocity{0, 1, 0})
i = 1:    Spawn2(Position{1, 0, 0}, Velocity{0, 1, 0})
i = 2:    Spawn2(Position{2, 0, 0}, Velocity{0, 1, 0})
...
i = N-1:  Spawn2(Position{N-1, 0, 0}, Velocity{0, 1, 0})
```

All entities live in one composition class — every entity carries
every queried component, in dense ID order. This is each backend's
happy path:

- **archetype**: one matching archetype, walked once per tick.
- **sparsesetslice**: per-row probe of the non-driver column always
  succeeds (the absence sentinel never fires).
- **sparsesetgroup**: declared owning group `{Position, Velocity}`
  matches the query; fast path walks the owned prefix
  `[0, g.size)` without per-row sparse-side work.

What it measures: raw single-archetype iteration throughput. The
discriminator across backends is the per-row work each one performs
when nothing fails — the sparse-side load-compare-load-store sequence
for sparsesetslice, the lockstep-direct walk for sparsesetgroup, the
per-archetype column-map indirection for archetype.

## multi_full and multi_partial

Two workloads sharing the same setup but differing in their queries.
The composition mixes entities across six classes cycled by
`i % 6`, chosen to fragment archetype's matching set into multiple
matching archetypes and to populate sparsesetslice's driver column
with rows that fail the per-row probe.

Spawn pattern (six composition classes, cycled):

```
i % 6 == 0:  Spawn(Position, Velocity, Health, Tag)
i % 6 == 1:  Spawn3(Position, Velocity, Health)
i % 6 == 2:  Spawn3(Position, Velocity, Tag)
i % 6 == 3:  Spawn2(Position, Velocity)
i % 6 == 4:  Spawn3(Position, Health, Tag)
i % 6 == 5:  Spawn2(Velocity, Tag)
```

Cycling — rather than blocked spawning all class 0 first, then all
class 1, etc. — interleaves classes so each backend's data layout
sees realistic mixed input. Blocked spawning would let the prefetcher
win unrealistically on dense per-class prefixes.

Per-class component membership:

| class | P | V | H | Tag |
|-------|:-:|:-:|:-:|:---:|
| 0     | ✓ | ✓ | ✓ |  ✓  |
| 1     | ✓ | ✓ | ✓ |     |
| 2     | ✓ | ✓ |   |  ✓  |
| 3     | ✓ | ✓ |   |     |
| 4     | ✓ |   | ✓ |  ✓  |
| 5     |   | ✓ |   |  ✓  |

### multi_full

Query: `{Position, Velocity, Health}`. Per tick: add Velocity to
Position and decrement `Health.Current`. The Health write keeps the
third-component `Get` engaged against compiler dead-code-elimination
of the Health read.

Iterated subset: classes 0 and 1 — **1/3 of N entities**. Per
backend:

- **archetype**: walks two matching archetypes (`{P,V,H,Tag}` and
  `{P,V,H}`). The iterator's per-archetype matching cost is
  exercised across more than one matching archetype for the first
  time relative to iteration baseline.
- **sparsesetslice**: drives the Position column (5/6 of N). For
  each driver row, probes Velocity and Health. Of P-having rows, only
  2/5 also have both V and H — the other 3/5 fail one or both probes
  and skip via `continue NextRow`. The probe-and-skip cost is real
  even when the path skips.
- **sparsesetgroup**: declared owning group `{P, V, H}` matches the
  query; fast path walks the owned prefix `[0, g.size)` without
  sparse-side work. The owned prefix contains exactly the entities
  in classes 0 and 1 — the lockstep-on-Spawn discipline placed them
  there as they were spawned.

### multi_partial

Query: `{Position, Velocity}`. Per tick: add Velocity to Position
(same body as iteration baseline; only the entity population differs).

Iterated subset: classes 0, 1, 2, and 3 — **2/3 of N entities**. Per
backend:

- **archetype**: walks four matching archetypes (`{P,V,H,Tag}`,
  `{P,V,H}`, `{P,V,Tag}`, `{P,V}`). Per-archetype overhead amortizes
  over a wider matching population than `multi_full`'s.
- **sparsesetslice**: drives the Position column (5/6 of N). For
  each driver row, probes Velocity. Of P-having rows, 4/5 also have
  V — 1/5 fail the V probe and skip.
- **sparsesetgroup**: declared owning group `{P, V, H}` does **not**
  match the query (the query is a strict subset). Fallback path runs
  slice-style probes — functionally identical to sparsesetslice's
  iteration plus a small unified-iterator tax (per-Next mode-flag
  branch and per-row index writes the dual-mode iterator pays to
  keep `Next`/`Entity`/`Get` coherent across both modes).

## Group declarations

Each workload's `*Groups()` function returns the owning groups passed
to `sparsesetgroup.New(alloc, groups)` at construction. The harness
wires these in through `selectWorkloadGroups`.

| Workload      | Declared groups                  |
|---------------|----------------------------------|
| iteration     | `{Position, Velocity}`           |
| multi_full    | `{Position, Velocity, Health}`   |
| multi_partial | `{Position, Velocity, Health}`   |

`multi_full` and `multi_partial` share a group declaration so
sparsesetgroup's setup is identical between the two scenarios; only
the queried set differs. This isolates the fast-path-vs-fallback
distinction from setup-cost variations — the lockstep-on-Spawn
work happens identically in both cases, and the timed tick is the
only thing that changes.
