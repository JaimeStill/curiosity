# Render thread (concept)

## Question

What threading model does rendering require — a single render thread,
parallel command recording, or something else — and how does that
model compose with the inner tier's shared threading commitment? Within
that model, who holds GPU resources and who submits commands?

## Constraints already locked in

- Rendering is inner-tier (D-002; `design/engine/runtime.md` —
  Inner-tier members).
- Rendering is the most demanding hot-path consumer; meshing,
  visibility, and command submission all run within the frame budget
  (`design/engine/runtime.md` — Inner-tier members).
- Rendering's threading model is a load-bearing inner-tier commitment,
  not an implementation detail downstream of other members
  (`design/engine/runtime.md` — Inner-tier members).
- Inner-tier members share their threading model (D-002;
  `design/engine/runtime.md` — Inner-tier members section opener); the
  scheduler participates in that commitment
  (`concepts/engine/scheduler.md`).
- The runtime owns inner-tier resource lifetime, GPU resources included
  (`design/engine/runtime.md` — Resource ownership).
