# Rendering primitives (concept)

## Question

What set of rendering primitives does the engine natively support, and
how do they layer to produce a frame? Voxels are the spatial substrate;
what complementary primitives — GPU particles, procedural meshes,
billboards and impostors, signed distance fields, or others — handle
content voxels handle poorly, and how do they share the inner tier with
voxel data and the ECS?

## Constraints already locked in

- Rendering is inner-tier (D-002; `design/engine/runtime.md` —
  Inner-tier members).
- Rendering produces the frame's image and is the most demanding
  hot-path consumer; meshing, visibility, and command submission all
  run within the frame budget (`design/engine/runtime.md` —
  Inner-tier members, Rendering paragraph).
- Voxel data is its own inner-tier member alongside ECS and rendering,
  not represented as entities (`design/engine/runtime.md` —
  Inner-tier members, Voxel data paragraph).
- Inner-tier members share memory layout, threading model, and frame
  lifecycle (D-002; `design/engine/runtime.md` — Inner-tier members
  section opener); a new dense substrate that joins the inner tier
  inherits that shared commitment.
- Rendering's threading model is a load-bearing inner-tier commitment,
  not an implementation detail downstream of other members
  (`design/engine/runtime.md` — Inner-tier members, Rendering
  paragraph; `concepts/engine/render-thread.md`).
