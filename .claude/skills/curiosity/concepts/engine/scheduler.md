# Scheduler (concept)

## Question

What shape of declaration do inner-tier members expose for the data
they read and write each frame, and what algorithm does the runtime
apply to that declaration to order work and decide what runs in
parallel?

## Constraints already locked in

- The runtime owns scheduling; members do not run their own loops
  (`design/engine/runtime.md` — Scheduling, Lifecycle).
- Members declare data needs and dependencies; the runtime orders the
  work (`design/engine/runtime.md` — Scheduling).
- Inner-tier members share memory layout, threading model, and frame
  lifecycle (D-002; `design/engine/runtime.md` — Inner-tier members).
  The scheduler's threading model is part of that shared commitment,
  not a per-member choice.
- All inner-tier work for a frame runs inside that frame's scheduler
  graph (`design/engine/runtime.md` — Tier criterion).
- The engine's unit of scheduled work — what declares data needs and
  cadence — is formally named "Task" (D-017). The task declaration
  surface encodes read/write per component, direct vs. deferred
  mutation, and task type with subtypes per-frame, fixed-step,
  every-N, and conditional (D-016). The scheduler builds its graph
  from this metadata: read/write determines parallelism; direct vs.
  deferred determines which mutations land in-task versus through the
  runtime's command buffer; task type determines which schedule the
  task is enrolled in. Events are not encoded at this surface — they
  remain a runtime-provided primitive (channels with writers and
  readers) any task can use, with concrete shape deferred to
  experimentation (D-016).
