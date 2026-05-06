# sparsesetslice

Plain sparse-set with slice-backed sparse mappings — the active
sparse-set carry-forward backend in this experiment. One of four
backends; see the experiment [README](../README.md) for the
comparison's question and shared interface. The map-sparse cousin
lives in `sparsesetmap/` as the implementation behind the
iteration-baseline measurement comparison.

## What's in a column

Each component type gets one `column`. Component values are stored
back-to-back as raw bytes — the storage layer is type-erased, so a
column tracks the byte width of one value (`size`) and packs N values
into a flat `dense []byte` buffer.

For Position (3× float32 = 12 bytes), holding three entities, a
column looks like:

```
Position column
  size:     12

  entities: [ 1 ,  2 ,  3 ]
  dense:    │ pos@1 │ pos@2 │ pos@3 │           (3 × 12 = 36 bytes)
  sparse:   indexed by EntityID → dense row
              sparse[0] = -1   (sentinel: no entry)
              sparse[1] =  0
              sparse[2] =  1
              sparse[3] =  2
```

Three parallel views of the same data:

- `entities[i]` — the EntityID at row `i`.
- `dense[i*size : (i+1)*size]` — the value bytes for row `i`.
- `sparse[id]` — the row index for `id`, or `-1` if the column
  doesn't hold a component for that entity.

Each column is independent of every other. There is no cross-column
ordering invariant: row `i` in Position and row `i` in Velocity may
or may not refer to the same entity, depending on the order spawns
arrived.

## How iteration works

Query takes a component-set and builds a `queryRef` per requested
component, each carrying a column pointer and a row-index field. It
picks the smallest-cardinality column as the **driver**. Iteration
walks the driver's `entities`; each row, every non-driver column gets
a sparse probe to discover the entity's row in that column.

For a query on `{Position, Velocity}` with both columns holding
entities 1, 2, 3 (equal sizes; driver defaults to `refs[0]`):

```
walk row = 0 .. len(driver.entities)-1
  row 0: entity = driver.entities[0] = 1
         refs[0].index = 0                       (driver shortcut)
         Velocity.sparse[1] = 0  → refs[1].index = 0
         yield (Position dense at row 0, Velocity dense at row 0)
  row 1: entity = 2
         refs[0].index = 1
         Velocity.sparse[2] = 1  → refs[1].index = 1
         yield
  row 2: entity = 3
         refs[0].index = 2
         Velocity.sparse[3] = 2  → refs[1].index = 2
         yield
```

The non-driver probe is the structural cost of this strategy. Each
non-driver column at each row pays:

- a slice-header load (`refs[i].col.sparse`),
- a bounds check (`int(entity) < len(sparse)`),
- a sentinel test (`sparse[entity] >= 0`),
- a final index load (`int(sparse[entity])`).

For the iteration baseline these all return the row index unchanged
(every entity carries both components, so the sentinel never fires)
— the load-compare-load-store sequence still runs at every row. This
is exactly the work the `sparsesetgroup/` fast path skips when a
query matches a declared owning group.

When the sparse probe returns `-1`, or `EntityID >= len(sparse)`, the
row is skipped via `continue NextRow` on the outer labeled loop —
that entity isn't in the non-driver column. Iteration continues from
the next driver row.

## Mixed populations

If entity 12 carries Position but not Velocity:

```
Position column                  Velocity column
  entities: [10, 11, 12]           entities: [10, 11]
  dense:    │10│11│12│             dense:    │10│11│
  sparse:   10→0, 11→1, 12→2       sparse:   10→0, 11→1, 12→-1
```

A query on `{Position, Velocity}`: driver is Velocity (smaller column
wins). Walk Velocity's two entities, probe Position for each. Entity
12 is never reached — it's not in Velocity's `entities`.

If the driver were Position (forced, e.g. by query order or
equal-size tie), row 2 would probe Velocity's `sparse[12]`, hit the
sentinel `-1`, and skip via `continue NextRow`. Same outcome, more
work.

The driver-picks-smallest rule keeps the outer loop bounded by the
narrowest column. It's a defensive minimum-iteration choice, not a
correctness requirement.

## Spawn

Allocate a new EntityID; for each component in the spawn:
get-or-create the column, ensure sparse capacity for the new id, set
`sparse[id] = len(entities)`, append the id to `entities`, append
the value bytes to `dense`. No cross-column coordination — each
column is independent.

Simpler than the group variant's Spawn (no lockstep maintenance) and
simpler than archetype's Spawn (no archetype lookup, no row
tracking). This backend trades cheap mutation for the slightly more
expensive iteration above.

## Sparse mapping — slice with sentinel

This backend's defining choice: `sparse` is `[]int32` indexed by
EntityID, with `-1` as the absence sentinel. The cousin
`sparsesetmap/` uses `map[storage.EntityID]int` instead and is
otherwise identical in shape. The slice form is faster because the
per-probe cost is a direct array load (bounds check + slice index)
versus the map's hash function plus bucket-table indirection;
sequential-ID iteration also stays prefetcher-friendly with the slice
form (three concurrent sequential streams the prefetcher handles
trivially).

Memory cost is O(max EntityID), not O(entity count). That matters
once Despawn lands and recycled-vs-monotonic ID strategy comes into
play; see `concepts/engine/entity-allocator.md`.
