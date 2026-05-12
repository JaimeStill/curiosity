# Outer-tier contract (concept)

## Question

What is the shape of the contract that outer-tier members plug into?
What surfaces does it expose, what flows across them, and what
guarantees does the runtime offer about timing, threading, and resource
lifetime when an outer-tier member is invoked or invokes the runtime?

## Constraints already locked in

- Outer-tier members plug into the runtime through a stable contract
  (D-002; `design/engine/runtime.md` — Outer-tier members section
  opener).
- The contract is stable enough that the runtime can evolve inner-tier
  internals freely as long as it holds, and outer-tier members can be
  substituted, layered, or absent without forcing inner-tier rework
  (`design/engine/runtime.md` — Outer-tier members section opener).
- Outer-tier members produce output on cadences other than the frame —
  audio rate, network tick, I/O completion, user input
  (`design/engine/runtime.md` — Outer-tier members section opener).
- The runtime brokers outer-tier resources; their lifetime is reached
  through the contract, not managed by the runtime directly
  (`design/engine/runtime.md` — Resource ownership).
- Each outer-tier member's interaction with the inner tier is named in
  `design/engine/runtime.md` — Outer-tier members (audio: one-way event
  stream; UI: composited draw plus input routing; storage: async I/O
  with runtime-brokered consistency; networking: runtime-brokered
  boundary; content pipeline: resource side of the contract). The
  contract must support each of these without dragging inner-tier
  internals into its surface.
- The contract is asymmetric. Each outer-tier member exposes its own
  unique outbound surface to the runtime, reflecting its role and
  cadence (audio: event-consumption-shaped; storage: async-I/O-shaped;
  UI: bidirectional). The runtime's *inbound* surface to outer-tier
  members — the API outer-tier code uses to read or write engine state
  — is consistent across members. Whether implemented as a single
  handle or as a coherent set of handles, every outer-tier member sees
  the same set of capabilities; no member receives a tailored sub-API.
  This refines the stable-contract commitment in D-002; promotion to a
  discrete decision is deferred until the contract receives a depth
  pass.
- Per-concern contracts (the outbound interface plus the inner-tier-side
  bridge code that consumes it) live in per-concern packages within
  the engine module per D-028: `engine/audio/`, `engine/network/`,
  `engine/storage/`, `engine/content/`. Interface-at-consumer
  (conventions §5) places each contract alongside the bridge code
  that consumes it; a centralized `engine/outer/` umbrella was
  considered and rejected during D-028 for fracturing that cohesion.
- Lifecycle binding is uniform across concerns. Each per-concern
  contract includes a `Start(*lifecycle.Coordinator) error` method
  per D-028; the Coordinator (in `engine/lifecycle/`, lifted from
  herald's `pkg/lifecycle/`) orchestrates startup, readiness, and
  shutdown across all outer-tier modules. The plug-in mechanism is
  the Coordinator; the per-concern outbound surface is the bespoke
  shape the asymmetry framing names.

## Open design surface

With the lifecycle binding settled by D-028, the concept's remaining
forward-looking work is the **frame-loop adaptation**: how the
runtime's tick interacts with outer-tier members whose cadences are
independent of the frame (audio rate, network tick, I/O completion,
user input). Herald's Coordinator was shaped for an HTTP request
loop where members react to inbound requests; the engine drives a
frame loop and brokers cross-cadence boundaries on the outer-tier
members' behalf. The shape of that brokering — when the runtime
yields to outer-tier consumers, how snapshots cross the boundary
without forcing inner-tier internals through the contract — is the
concept's primary remaining surface.
