# Animation approaches (concept)

## Question

What animation approaches does the engine natively support, and where
in the inner tier do they live? The candidates surfaced — procedural
animation (IK-driven), voxel-rig transforms with simple hierarchical
rotations, vertex-shader deformation, particle-formed entities, and
rigid-transform-only — span a wide range of implementation cost and
fidelity. Which does the engine offer as first-class capabilities, and
what surface do they share with rendering, physics, and the ECS?

## Constraints already locked in

- Rendering consumes per-entity transforms and per-frame visual state
  (`design/engine/runtime.md` — Inner-tier members, Rendering
  paragraph).
- Physics writes updated transforms back into the ECS each frame
  (`design/engine/runtime.md` — Inner-tier members, Physics paragraph);
  animation interacts with the same transform surface and must
  coordinate with physics in the scheduler graph.
- Inner-tier members share memory layout, threading model, and frame
  lifecycle (D-002; `design/engine/runtime.md` — Inner-tier members
  section opener); animation, wherever it lives in the inner tier,
  inherits that commitment.
- The set of rendering primitives the engine supports
  (`concepts/engine/rendering-primitives.md`) constrains which
  animation approaches are needed: voxel-rig animation is only
  meaningful if voxel-shaped entities exist as a primitive;
  particle-formed animation only if particles are first-class.
