# Engine Runtime

The runtime is the inner-tier host. It owns the frame loop, schedules
inner-tier work, and brokers the contract that lets outer-tier
components plug in without reaching into inner-tier internals. This
document is forward-looking reference material — claims grounded in
`history/decisions.md` (especially D-002), in source code, or in hard
external constraints. Material that is not yet settled lives under
`concepts/engine/`, not here.

## Tier criterion

The two-tier shape (D-002) draws its line at hot-path coupling. A
subsystem is inner-tier when its per-frame work shares memory layout,
threading assumptions, or lifecycle phasing with other inner-tier
subsystems closely enough that a stable interface between them would
become the bottleneck. A subsystem is outer-tier when its work is
cold-path relative to the frame loop and a stable contract costs
little to honor.

The test, applied to a candidate subsystem:

- Does it touch per-entity or per-voxel data on the critical frame
  path? → inner.
- Must it interleave with other inner-tier work inside a single
  frame's scheduler graph? → inner.
- Can it operate against a stable interface without forcing inner-tier
  internals to leak through that interface? → outer.

A subsystem that flips from outer to inner is a workflow event, not
just an implementation event: the move is recorded in
`history/decisions.md` and the relevant inner-tier internals are
revisited rather than wedged through the existing contract.

## Inner-tier members

The inner tier shares its memory layout, its threading model, and its
frame lifecycle. Members do not communicate through interfaces sized
to absorb arbitrary substitution; they communicate through the shared
frame and the data already in place. That coupling is the point — it
is what keeps interfaces from becoming bottlenecks on the hot path.

**ECS** stores entities and the components attached to them, and
drives the per-frame iteration that inner-tier work runs over. Its
storage layout is the substrate every other inner-tier member reads
from or writes to; the choice of layout
(`concepts/engine/ecs-storage.md`) ripples through physics and
rendering.

**Voxel data** holds the world's voxels and the access patterns that
read or modify them. Both meshing (a rendering concern) and collision
queries (a physics concern) reach into it on the hot path; mutations
must coordinate with the frame's scheduler so that consumers see a
consistent snapshot of the world they are simulating or drawing.

**Physics** integrates motion and resolves contacts each frame,
reading entity transforms and voxel data and writing updated
transforms back into the ECS. Routing those inputs through a stable
contract would impose data-shape constraints on ECS and voxel data
that the inner tier is built to avoid; physics belongs alongside its
inputs rather than across an interface from them.

**Rendering** produces the frame's image from ECS state and voxel
data. It is the most demanding hot-path consumer — meshing,
visibility, and command submission all run within the frame budget —
and the threading model it requires
(`concepts/engine/render-thread.md`) is one of the inner tier's
load-bearing commitments rather than an implementation detail
downstream of the others.

## Outer-tier members

The outer tier is cold-path relative to the frame loop. Members
produce their output on cadences other than the frame — audio rate,
network tick, I/O completion, user input — and they plug into the
runtime through a stable contract whose specifics live in
`concepts/engine/outer-tier-contract.md`. Stability is the trade: the
runtime can evolve inner-tier internals freely as long as the contract
holds, and outer-tier members can be substituted, layered, or absent
without forcing inner-tier rework.

**Audio** renders sound from events emitted by the inner tier or by
game code. It runs on its own timeline (audio rate, separate from
frame rate), and its only contact with the frame is a one-way event
stream; nothing the frame produces depends on audio's response.

**UI** draws interface chrome and routes input events back to the
game. Although it draws, its draw is composited atop the rendered
frame and does not need to interleave with inner-tier scheduling.

**Storage** persists and loads world state, save data, and
configuration. It is I/O-bound and asynchronous; the runtime brokers
in-flight work so that inner-tier consumers see consistent state when
they read.

**Networking** sends and receives game state across whatever
transport the game requires. Its cadence is set by network conditions
rather than the frame, and the runtime brokers the boundary through
the outer-tier contract rather than letting inner-tier subsystems
reach across it directly.

**Content pipeline** imports and processes assets — meshes, textures,
voxel models, audio data — into runtime-ready forms. Its work is
offline or asynchronous, and it feeds the inner tier through the
resource side of the outer-tier contract.

## Runtime roles

The runtime is the seat of control for everything in this document.
Inner-tier members run inside it; outer-tier members plug into it.
This section names the runtime's responsibilities at the depth this
document operates — what the runtime owns, not how it does it.
Specifics for each role live in `concepts/engine/` and graduate to
design only when their commitments are firm.

**Frame loop.** The runtime drives the per-frame tick: it advances
time, dispatches inner-tier work, and yields to outer-tier consumers.
Frame structure beyond "the runtime owns it" is concept-tier.

**Scheduling.** Within a frame, the runtime decides what runs when
and what runs in parallel. Members declare their data needs and
dependencies; the runtime orders the work. The shape of that
declaration and the scheduler's algorithm live in
`concepts/engine/scheduler.md`.

**Lifecycle.** The runtime owns startup, shutdown, and the phase
transitions between them. Members register with the runtime; the
runtime drives them, not the reverse.

**Resource ownership.** The runtime owns the lifetime of inner-tier
resources — ECS storage, voxel data buffers, GPU resources — and
brokers access for the members that need them. Outer-tier resources
are reached through the outer-tier contract; the runtime does not
manage their lifetime directly.
