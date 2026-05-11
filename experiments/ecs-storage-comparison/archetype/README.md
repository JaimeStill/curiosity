# archetype

Archetype-based storage — entities grouped by exact component-set
into self-contained dense tables. One of four backends in this
experiment; see the experiment [README](../README.md) for the
comparison's question and shared interface.

## Strategy

Every entity has a component-set — the components attached to it.
Two entities with the same component-set live in the same
**archetype**, sharing dense storage. Two entities with different
component-sets live in different archetypes, even when their
component-sets overlap heavily. When an entity gains or loses a
component, its archetype changes and the row moves.

Iteration walks every archetype whose component-set is a *superset*
of the query — those archetypes have all the components the query
asks for, plus possibly more. The per-row work is dense column
indexing only: no sparse probe, no absence test, because every
entity in a matching archetype has every queried component by
construction.

## What's in an archetype

For three entities all carrying both Position and Velocity, the
picture is one archetype:

```
archetype:  signature = {Position, Velocity}

  cids:     [Position, Velocity]   (sorted)
  entities: [ 1 ,  2 ,  3 ]

  columns:
    Position column
      size: 12, data: │ pos@1 │ pos@2 │ pos@3 │     (3 × 12 = 36 bytes)
    Velocity column
      size: 12, data: │ vel@1 │ vel@2 │ vel@3 │     (3 × 12 = 36 bytes)
```

A column carries `cid`, `size`, and a `data []byte` buffer. Notably
absent compared to the sparse-set variants:

- No per-column `entities` slice. The archetype owns one `entities`
  list shared by all columns.
- No `sparse` mapping. There is no per-entity row-index lookup at
  the column level; the archetype's `entities` slice and `data`
  buffers are aligned by row, and `entities[i]` is the entity at
  row `i` of every column in the archetype.

Cross-column row alignment is the archetype's defining invariant.

## Multiple archetypes

Suppose entity 4 spawns carrying Position, Velocity, *and* Health.
That's a different component-set, so it lives in a new archetype:

```
archetype A:  signature = {Position, Velocity}
  entities: [1, 2, 3]
  Position data: │1│2│3│            Velocity data: │1│2│3│

archetype B:  signature = {Position, Velocity, Health}
  entities: [4]
  Position data: │4│                Velocity data: │4│
  Health data:   │4│
```

Each archetype carries its own dense storage. Entity 4's Position is
not in archetype A's Position data — it's in archetype B's Position
data. The two archetypes are independent and unaware of each other
at the data level.

## Storage

Top-level state:

- `archetypes map[component.Signature]*archetype` — every archetype
  the backend has seen, keyed by its signature.
- `locations []location` — entity → (archetype pointer, row),
  indexed directly by `int(id)`. Sentinel `arch == nil` marks empty
  slots; `slices.Grow` expands capacity when a spawned entity's ID
  exceeds current length. Reuses the dense-ID guarantee from the
  entity allocator's free-list recycler — the same assumption
  sparsesetslice's `[]int32` sparse mappings carry. Despawn-apply,
  Read, and Write can find any entity's data without scanning
  archetypes (D-026).
- `alloc *entity.Allocator` — shared entity-ID allocator (free-list
  recycler). Spawn calls `alloc.Allocate()`; Despawn calls
  `alloc.Free(id)` after the row's swap-remove completes.

`component.Signature` is a `uint64` bitmask — each `component.ID`
is a bit position. `Set(cid)` flips the bit; `Contains(other)` is a
bitmask superset check. The 64-bit width caps component types at 64,
which suffices for this experiment.

## How iteration works

Query takes a component-set, computes its signature, and walks every
archetype testing `archetype.signature.Contains(querySig)` —
archetypes whose component-sets are supersets of the query. Matching
archetypes go into the iterator's `matches` slice.

Iteration proceeds archetype by archetype:

```
query: {Position, Velocity}
matches: [archetype A {P,V}, archetype B {P,V,H}]

step    arch  row    entity   resolves to
  1      A     0      1       A.Position data[0..12), A.Velocity data[0..12)
  2      A     1      2       A.Position data[12..24), A.Velocity data[12..24)
  3      A     2      3       A.Position data[24..36), A.Velocity data[24..36)
  4      B     0      4       B.Position data[0..12), B.Velocity data[0..12)
                              (B.Health is in archetype B but not asked for)
```

When one archetype's entities are exhausted, the iterator advances
`index` to the next archetype and resets `row` to 0. `Get(cid)`
resolves the current archetype's column via `arch.columnFor(cid)` —
a linear scan over the archetype's `cids` slice to find the column
matching the requested ComponentID.

The per-row work is a single column lookup (short scan over `cids`)
plus byte arithmetic for the dense offset. No sparse probe, no
absence test — the archetype's component-set guarantees presence.

## Spawn

For each spawn:

1. Compute the spawn's signature (`sig.Set(cid)` per component).
2. Find or create the archetype for that signature
   (`getOrCreateArchetype`). New archetypes get sorted `cids` and
   one column per component with the size pre-resolved from
   `component.TypeOf(cid).Size()`.
3. Mint a new entity ID via `s.alloc.Allocate()`.
4. Append the entity to the archetype's `entities`; append each
   component's bytes to its column's `data`.
5. Record the entity's location as `(archetype, row)` for later
   lookup.

Per-spawn cost decomposes into the signature computation (one bit
set per component), the archetype-table lookup (`archetypes` map
keyed on signature, O(1) average), the column-data appends (one per
component), and the location-slice write (direct index into
`s.locations[id]`, with occasional `slices.Grow` capacity expansion
when an ID exceeds current length). All four are constant work for
a fixed component count and dense-ID input.

Within a single archetype, Spawn is a simple dense append. Structural
mutation that *crosses* archetype boundaries — when Attach or Detach
land — moves a row from the old archetype's columns to the new
archetype's columns and updates the location entry. That's where the
strategy pays its iteration ergonomics back: attach/detach churn is
structurally more expensive than sparse-set, where each column is
independent and stays in place.
