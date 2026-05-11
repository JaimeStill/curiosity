# Entity allocator (concept)

## Question

How does the engine allocate entity IDs, recycle them when entities
despawn, and prevent stale handles from silently aliasing recycled
entities? The decision is layered above the ECS storage strategy
(`concepts/engine/ecs-storage.md`) — every storage backend benefits
from the same allocator — but its choices ripple into entity-ID
representation, the cost profile of Spawn/Despawn, and the
correctness of any system that holds an EntityID across frames.

## Constraints already locked in

- The runtime owns ECS storage lifetime (`design/engine/runtime.md`
  — Resource ownership). Entity allocation sits inside that.
- Inner-tier members share memory layout (D-002). EntityID is part
  of that shared layout: whatever shape the allocator chooses for
  IDs becomes the shape inner-tier members read and write.
- ECS storage strategy is sized for entity-and-component data
  (`concepts/engine/ecs-storage.md`); voxel data and possibly
  particles live outside ECS. The entity allocator scopes to ECS
  entities; non-ECS dense data has its own identity primitives.

## Open design questions

- **Recycling queue.** Despawn places the freed ID into a reserve
  pool; Spawn dequeues a recycled ID before allocating fresh.
  Without recycling, any storage layout indexed by EntityID
  (especially slice-based sparse-set per the
  `experiments/ecs-storage-comparison/` findings) grows unboundedly
  with cumulative spawns. With recycling, memory cost stabilizes at
  peak-concurrent-entity-count rather than cumulative-spawn-count.

- **Generation counters for stale-handle protection.** Recycling
  creates an ABA hazard: code holding a reference to "entity 42"
  from before despawn would silently alias a different entity once
  42 is recycled. Standard fix is to pair each ID with a generation
  counter that increments on recycle. EntityID becomes
  `(index, generation)` packed into a single integer (Bevy and EnTT
  both pack 32+32 into 64 bits). Storage keys on the index portion;
  the generation is validated at every API call site that takes an
  EntityID and rejected as a stale handle if it doesn't match.

- **Compaction over timeouts for memory reclamation.** When the
  reserve queue grows large relative to live entities and storage
  arrays indexed by EntityID have correspondingly grown, reclaiming
  that memory needs an explicit compaction operation that reassigns
  live entity IDs to the low end of the ID space. Timeouts
  (expiring queued IDs) reintroduce monotonic ID growth, couple
  allocation to wall-clock time (which is layer-mixing), and produce
  nondeterministic behavior bad for debugging and reproducibility.
  Compaction is structural — triggered by an explicit policy (e.g.,
  max ID > 4× live count) at a frame boundary safe for ID
  reassignment.

- **Allocator above storage layer.** Entity ID allocation is a
  separate concern from component storage. All storage backends
  benefit from one recycling allocator; iteration code in storage
  backends should not be entangled with free-list logic. The
  layering supports independent test/swap of allocator policies
  without touching storage code, and keeps storage backends
  comparable on equal terms (each measured against the same
  allocator behavior rather than its own private ID-allocation
  shortcut).
