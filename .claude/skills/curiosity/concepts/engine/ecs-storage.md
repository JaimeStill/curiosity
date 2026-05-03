# ECS storage (concept)

## Question

What storage strategy does the ECS use for components, and what
iteration patterns must that strategy serve so that physics, rendering,
and the scheduler can read and write efficiently against it each frame?

## Constraints already locked in

- ECS is inner-tier (D-002; `design/engine/runtime.md` — Inner-tier
  members).
- The ECS stores entities and components and drives per-frame iteration
  that inner-tier work runs over (`design/engine/runtime.md` —
  Inner-tier members, ECS paragraph).
- The chosen storage layout ripples through physics and rendering; it
  is not a private ECS implementation detail (`design/engine/runtime.md`
  — Inner-tier members, ECS paragraph).
- Inner-tier members share memory layout (D-002;
  `design/engine/runtime.md` — Inner-tier members section opener); the
  ECS is the substrate that shared commitment is expressed against.
- The runtime owns ECS storage lifetime (`design/engine/runtime.md` —
  Resource ownership).
- Some forms of dense data live outside the ECS rather than as
  entities and components: voxel data is already a separate inner-tier
  member (`design/engine/runtime.md` — Inner-tier members, Voxel data
  paragraph); particle data may be another at scale
  (`concepts/engine/rendering-primitives.md`). ECS storage strategy is
  sized for entity-and-component data, not for arbitrary dense data.
