# sparsesetmap

Sparse-set storage with map-backed sparse mappings. Structurally
identical to `sparsesetslice/` — same column shape, same iteration
algorithm, same Spawn — except `sparse` is a
`map[storage.EntityID]int` instead of `[]int32` with sentinel. This
variant is preserved as the implementation behind the
iteration-baseline measurement comparison in the experiment
[README](../README.md)'s 2026-05-06 Finding; per R-007 it is not
built out further. Production sparse-set engines (EnTT and similar)
use paged sparse arrays rather than hash maps, and the map variant's
job here was informational — once it had quantified the
sparse-mapping representation's contribution to iteration cost,
further build-out wouldn't have yielded new signal.

See [`sparsesetslice/README.md`](../sparsesetslice/README.md) for the
data model and algorithm walkthroughs. This document covers only
what makes the map variant different.

## What changes from the slice variant

A column's `sparse` field is a Go map instead of a slice:

```
type column struct {
    cid      component.ID
    size     uintptr
    sparse   map[entity.ID]int
    dense    []byte
    entities []entity.ID
}
```

Two consequences follow from that change.

**Iteration.** The per-row non-driver probe is a map lookup instead
of a bounds-and-sentinel test on the sparse slice:

```
slice variant:                          map variant:
  if int(e) >= len(sparse) ||             index, present := sparse[e]
     sparse[e] < 0:                       if !present:
      continue NextRow                        continue NextRow
  refs[i].index = int(sparse[e])        refs[i].index = index
```

**Spawn.** The slice variant's `ensureSparseCapacity` grow-and-fill
loop becomes a direct `sparse[id] = ...` map write; Go's runtime
handles bucket allocation.

Everything else — the column's `dense` and `entities`, the
iterator's `queryRef` structure, the driver-picks-smallest rule, the
rest of Spawn — is identical to the slice variant.

## What the variant surfaced

The map-vs-slice axis isolated the sparse-mapping representation's
contribution to iteration cost. In the iteration-baseline
measurements at 1k/10k/100k, sparsesetslice tracks archetype within
~1%, while sparsesetmap shows cache-cliff growth (8.94 → 18.87 →
23.95 → 31.35 ns/entity at the ascending scales) — the map's hash
function randomizes the access pattern and defeats prefetcher
behavior at scales where the working set spills out of L1. The
algorithmic shape is identical between the two variants; only the
per-probe load pattern diverges. See the experiment README's
2026-05-06 Finding for full data and interpretation.
