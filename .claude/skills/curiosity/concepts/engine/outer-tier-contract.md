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
