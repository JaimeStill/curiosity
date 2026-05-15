# Scheduler (concept)

## Question

What algorithm does the runtime apply to inner-tier members'
data-need declarations to order work and decide what runs in
parallel?

## Task declaration surface

The engine's unit of scheduled work is named **Task**. The naming
diverges from production-ECS-framework convention (Bevy's "System",
Unity DOTS's "ComponentSystem", flecs's "system") in favor of
alignment with the author's pre-existing service-architecture
mental model (queries, commands, tasks, events), where "task" maps
cleanly to a unit of scheduled work. "System" is reclaimed for
engine-level concerns (audio system, storage system, etc.) and for
the S in ECS — Go's package namespacing handles the ambiguity
(`ecs.System` and `audio.System` are distinct types). The
translation cost for readers familiar with Bevy or DOTS is mild and
acceptable given the project's posture as a learning artifact whose
primary reader is the author for the foreseeable future.

The Task declaration surface encodes three things:

1. **Read/write per component type.** Tasks declare which
   components they read and which they write; the runtime uses these
   declarations to prove which tasks can run in parallel. Read/write
   at the API is unconditionally required for safe parallel
   scheduling — without it, the runtime cannot prove which tasks can
   run concurrently against shared component data.

2. **Direct vs. deferred mutation.** Direct access (a borrowed
   reference to component data) for data mutation; a deferred
   command-buffer-style queue for structural mutation (spawn,
   despawn, attach, detach). The split is forced by the storage
   layer's iteration constraints regardless of which storage
   strategy is chosen — structural mutations must be deferred to
   safe sync points to preserve iterator validity. Encoding the
   split at the API makes mistakes impossible rather than merely
   strongly discouraged.

3. **Task type.** Every task declares its cadence — per-frame,
   fixed-step, every-N, conditional. The runtime builds the
   schedule from task type plus read/write declarations: read/write
   determines parallelism; direct vs. deferred determines which
   mutations land in-task versus through the runtime's command
   buffer; task type determines which schedule the task is enrolled
   in. Cadence is a property of *when* a task runs, not a base
   class of *what kind* of task it is — the production-ECS
   anti-pattern of many task class hierarchies (Unity DOTS-style)
   is avoided.

Events are *not* encoded at this surface. They remain a
runtime-provided primitive (channels with writers and readers) that
any task can use, but no special task type wraps them. Concrete
event shape (auto-clear vs. broadcast, observer-style vs. polled,
queued vs. reactive) is deferred to experimentation rather than
committed up front — production frameworks differ on this and the
choice is better made once the engine has enough use cases to
inform it.

## Constraints already locked in

- The runtime owns scheduling; members do not run their own loops
  (`design/engine/runtime.md` — Scheduling, Lifecycle).
- Members declare data needs and dependencies; the runtime orders
  the work (`design/engine/runtime.md` — Scheduling).
- Inner-tier members share memory layout, threading model, and
  frame lifecycle (`design/engine/runtime.md` — Inner-tier
  members). The scheduler's threading model is part of that shared
  commitment, not a per-member choice.
- All inner-tier work for a frame runs inside that frame's
  scheduler graph (`design/engine/runtime.md` — Tier criterion).
- The actor-model alternative for inner-tier subsystems
  (subsystem-per-goroutine with channel-mediated cross-subsystem
  communication) was considered and rejected. The tier criterion
  explicitly identifies stable interfaces between inner-tier
  subsystems as the bottleneck the inner tier exists to avoid;
  channels are stable interfaces with ~100 ns/op overhead. At
  inner-tier per-entity per-frame access patterns the channel
  arithmetic collapses the frame budget (~2.4M cross-subsystem
  ops/sec for physics↔ECS alone at 10k entities × 60 fps × 4
  ops/entity, before any actual work). Inner-tier multi-threading
  happens instead through (a) within-subsystem parallelism (worker
  pools over entity slices), (b) scheduler-aware parallel Tasks
  via the read/write declarations above, and (c) render-thread
  separation per `render-thread.md`. The actor model is the
  *outer-tier* pattern per `code/conventions.md` §10.
