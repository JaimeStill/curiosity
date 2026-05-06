# sparsesetgroup

EnTT-style opt-in owning-groups variant of the sparse-set storage
strategy. One of four backends in this experiment; see the experiment
[README](../README.md) for the comparison's question and shared
interface.

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
  sparse:   index by EntityID → dense row
              sparse[0] = -1   (sentinel: no entry)
              sparse[1] =  0
              sparse[2] =  1
              sparse[3] =  2
```

Three parallel views of the same data:

- `entities[i]` — the EntityID at row `i`.
- `dense[i*size : (i+1)*size]` — the value bytes for row `i`.
- `sparse[id]` — the row index for `id`, or `-1` if the column doesn't
  hold a component for that entity.

## What groups add

A group declares a component-set that participates in lockstep dense
ordering. The boundary index `g.size` separates the **owned prefix**
`[0, g.size)` — entities carrying every component the group declares —
from the **suffix** `[g.size, len(entities))` — entities in this
column but not eligible for the group.

For three entities all carrying both Position and Velocity, with the
group declared on `{Position, Velocity}`:

```
Position column                  Velocity column
  entities: [1, 2, 3]              entities: [1, 2, 3]
  dense:    │1│2│3│                dense:    │1│2│3│
  sparse:   1→0, 2→1, 3→2          sparse:   1→0, 2→1, 3→2

Group on {Position, Velocity}
  size:    3                       (every entity is in the prefix)
  columns: [&Position, &Velocity]
```

Read the lockstep invariant off the picture: row 1 in Position holds
entity 2; row 1 in Velocity also holds entity 2. Both columns agree
on every entry across the prefix.

## How iteration works

For a query that matches the group's set, iteration walks
`[0, g.size)`. Each row writes the dense index directly to every ref:

```
group fast path, query {Position, Velocity}:

  walk row = 0 .. g.size-1
    row 0:  refs[0].index = 0      (Position row 0 = entity 1)
            refs[1].index = 0      (Velocity row 0 = entity 1)
    row 1:  refs[0].index = 1
            refs[1].index = 1
    row 2:  refs[0].index = 2
            refs[1].index = 2
```

That's the entire per-row work. The lockstep invariant guarantees
`index == row` for every ref, so no sparse-side load is needed.

A plain sparse-set (`sparsesetslice/`) has to load `sparse[entity]`
for each non-driver column at every row to discover that column's
dense index — even when the answer is always the row. That's the
per-row work the group fast path skips. It's the source of the
measured iteration advantage in the experiment README's 2026-05-06
Finding.

For queries that don't match any declared group, iteration falls back
to slice-style: pick the smallest column as driver, walk its
entities, probe each non-driver column's sparse mapping per row, skip
rows where any non-driver column lacks the entity. Identical to the
slice variant.

## Mixed populations

The boundary index earns its keep when not every entity belongs to
the group. Picture three entities — 10 and 11 carrying both Position
and Velocity, 12 carrying only Position. Group still on
`{Position, Velocity}`:

```
Position column                  Velocity column
  entities: [10, 11, 12]           entities: [10, 11]
  dense:    │10│11│12│             dense:    │10│11│
  sparse:   10→0, 11→1, 12→2       sparse:   10→0, 11→1, 12→-1

Group on {Position, Velocity}
  size:    2                       (entity 12 is in the suffix)
```

`g.size` is 2, not 3. Position row 2 — entity 12 — sits in the suffix.
The fast path walks `[0, 2)` and yields entities 10 and 11 only.
Entity 12 is correctly skipped from the group iteration because it's
missing Velocity, and the lockstep invariant placed it past the
boundary.

For a single-component query on `{Position}` alone, the fast path
doesn't apply (declared group is `{Position, Velocity}`, not
`{Position}`). The fallback iterator walks Position's full range
`[0, 3)` and yields all three entities, including 12.

## Spawn — appending and the lockstep swap

A spawn runs in two phases.

**Phase A — append.** Allocate a new EntityID; for each component in
the spawn, append the value to the column. After Phase A, the new
entity sits at the tail of every column it touched.

Suppose we spawn entity 13 carrying both Position and Velocity into
the mixed-population picture above:

```
After Phase A:

Position column                  Velocity column
  entities: [10, 11, 12, 13]       entities: [10, 11, 13]
  dense:    │10│11│12│13│          dense:    │10│11│13│
  sparse:   10→0, 11→1,            sparse:   10→0, 11→1, 13→2
            12→2, 13→3                       12→-1

Group on {Position, Velocity}
  size:    2                       (Phase B hasn't run yet)
```

Position has entity 13 at row 3 — past the boundary. The group is
incoherent: 13 belongs in the prefix (it covers the group's set), but
it's sitting in the suffix.

**Phase B — lockstep swap.** For each participating column, swap row
`g.size` with row `tail`. In Position, swap row 2 (entity 12) with
row 3 (entity 13):

```
Position column after the swap:
  entities: [10, 11, 13, 12]
  dense:    │10│11│13│12│
  sparse:   10→0, 11→1, 13→2, 12→3
```

Three pieces of state move together: dense bytes (via
`swapDenseRows`), entity slot (`entities[row], entities[tail] =
entities[tail], entities[row]`), and the sparse mapping for both
displaced entities (`sparse[13] = 2`, `sparse[12] = 3`).

Velocity already had entity 13 at row 2, equal to its `g.size`, so
its swap is a self-swap — the `row == tail` short-circuit in Spawn
fires. After the per-column loop, `g.size` advances from 2 to 3.

The picture is coherent again:

```
Position column                  Velocity column
  entities: [10, 11, 13, 12]       entities: [10, 11, 13]
  dense:    │10│11│13│12│          dense:    │10│11│13│
  sparse:   10→0, 11→1,            sparse:   10→0, 11→1, 13→2
            13→2, 12→3                       12→-1

Group on {Position, Velocity}
  size:    3                       prefix: 10, 11, 13
                                   suffix: 12 (Position only)
```

Row `i ∈ [0, 3)` in either column gives the same entity. The fast
path resumes.

### Iteration-baseline degeneracy

In this experiment's iteration baseline, every entity carries the
declared group's full set. The suffix is always empty, and the
boundary always equals `tail` *before* the swap — the
`row == tail` short-circuit fires every time, and the swap path is
correct but never exercised. The algorithm is still shaped for the
general case so the measurement reflects what would ship, not a
workload-specific shortcut.

## Construction

`New(groups [][]ComponentID)` declares the owning groups up-front and
materializes the participating columns immediately. After `New`
returns, group state is fixed; Spawn maintains the lockstep invariant
from the first call forward, with no retroactive-materialization
path.

Declaration-at-construction matches what production owning-group
implementations (EnTT) do. Auto-declaring on first Query was
considered and rejected — it would require retroactive lockstep
materialization code with no production analogue, biased toward the
experiment's wiring rather than what a real engine looks like.

A `*Storage` constructed with no declared groups behaves entirely as
the slice variant; the fast path is never taken.

## Iterator — one struct, two modes

A single `iterator` carries both fast-path and fallback state with a
mode flag (`group *group`; nil = fallback). The shared shape lets
`Entity` and `Get` work identically across both modes — `Entity`
reads `refs[driver].col.entities[row]` (driver defaults to 0 in the
fast path; refs[0] is a group column whose entities[row] is valid by
the lockstep invariant), and `Get` reads `refs[i].index` (Next
maintains it in both modes — overwriting to `row` on the fast path,
resolving via sparse probe on the fallback).

This costs the fast path a small per-Next tax — the index-write loop
runs even though the fast path could otherwise expose `row` to Get
directly. A specialized fast-path iterator would elide those writes.
The unified design pays them as the cost of structural coherence
between modes; the measurement reflects what a general-purpose
shipped owning-group iterator would look like rather than a hand-
tuned fast path.
