---
name: curiosity
description: Workflow discipline for a generic Go-based voxel game engine, developed alongside a reference game that serves as a forcing function and validation target. Use whenever working in this repository - drafting or modifying design documentation, recording architectural decisions, performing context resets, evaluating new dependencies, or making decisions about engine architecture, components, or game systems. Always consult this skill before adding to design docs, taking on a dependency, or marking design context as resolved. The project favors learning through building, minimal dependencies, iterative development, and a single source of truth for every piece of context.
---

# Voxel Engine Project Workflow

This skill encodes the working agreement for the project. It is the entry point: the design documentation and history files referenced below carry the substance and are loaded as needed.

## Project Scope

The primary artifact is a generic Go-based voxel game engine, designed to support modern voxel-based games at high fidelity. The engine's design is not bound to any specific game.

A reference game is developed alongside the engine. Its role is to keep one concrete scenario sharp enough to surface real engine requirements rather than imagined ones, and to provide an aspirational target worth building toward. The reference game's premise lives in `design/game/premise.md`.

**The separation is load-bearing.** Engine decisions serve the generic voxel game case. When the reference game surfaces a need, it is generalized before being incorporated into engine design. Features that only make sense for the reference game live in game-side code, not in the engine. The reference game is one validation target among many possible games the engine should support.

This is as much a learning project as a building project. Decisions favor understanding over expedience.

## Working Principles

**Single source of truth.** Every piece of context exists in exactly one place: in source code or in design documentation, never both. A design doc section that describes implemented behavior is a defect.

**Plan only the next step.** No execution ordering, no roadmaps, no phases. The design documentation captures what will need to exist; the next concrete step is decided when the current step completes.

**Documentation decay.** Design docs only contain what code cannot express: the why behind decisions, alternatives considered and rejected, constraints not yet enforced, and intent for unimplemented work. Once code can fully express a design section, that section is removed and its absorption recorded in the reset log.

**Iterative depth.** Components and subsystems get shallow inventory entries first. Depth is added when work on that component is imminent, not before. Premature depth is the same trap as premature implementation.

**No premature optimization.** Correctness and clarity first. Performance work happens against measured problems, not anticipated ones.

**Vertical over horizontal.** When integrating multiple subsystems, build a deliberately crude end-to-end path before deepening any single component. Integration risk surfaces early or it surfaces expensively.

**Engine-first generalization.** When the reference game surfaces a requirement, the question is whether it generalizes to other voxel games. If yes, it informs engine design at the appropriate layer of abstraction. If no, it lives in game code. The engine never grows premise-specific concepts.

## Dependency Policy

Dependencies are taken on only when all three conditions hold:

1. **Active and well-supported** - recent commits, responsive maintainership, healthy issue activity.
2. **Open source** - no closed or source-available licenses.
3. **Critical to execution** - the project cannot reasonably proceed without it, or reimplementation would consume effort better spent on the project's core.

The standard library is preferred wherever it suffices. Convenience is not justification. Every new direct dependency is recorded in the decisions log with the reasoning that satisfied the three conditions; renewals or replacements supersede the prior entry.

Transitive dependencies are noted but not gated; direct dependencies are the surface where this discipline applies.

The dependency policy applies independently to engine and game code. Game-side dependencies do not flow into the engine; engine dependencies should be defensible without reference to game-specific needs.

## Repository Layout

```
SKILL.md                          this file
design/
  engine/
    runtime.md                    core engine runtime layout
    components.md                 inner and outer component inventory
    components/<name>.md          per-component depth, created on demand
  game/
    premise.md                    aspirational target for the reference game
    <other game design docs>      created as game design surfaces them
history/
  decisions.md                    append-only architectural decisions
  resets.md                       append-only reset transaction summaries
resources/
  assets.md                       curated asset acquisition resources
<source directories>              created as components materialize
```

Source layout reflects the same separation: engine code and game code live in distinct trees, and the engine has no compile-time dependency on game code.

The `resources/` directory holds reference material that is not subject to the documentation-decay discipline - curated lists, acquisition references, external pointers. It exists alongside design and history but follows different rules: entries are added or revised as the external landscape shifts, not as code absorbs them.

## Design Documentation Conventions

Design documents are reference material, not historical record. They describe the present forward-looking state of intent for things not yet built or not yet fully expressed in code.

### Engine design documents

**`design/engine/runtime.md`** is the load-bearing engine document. It defines the inner-tier boundaries (the systems that share memory and frame lifecycle: ECS, voxel data, physics, rendering), the outer-tier contract (the interface plug-in components implement), and the runtime's own responsibilities (frame loop, scheduling, lifecycle, resource ownership). Referenced by everything else and the most expensive to get wrong.

**`design/engine/components.md`** is shallow by default: a list of inner and outer components with one to two sentences each on responsibility and rough interface shape. Depth is added in `design/engine/components/<name>.md` only when that component is being worked on.

**`design/engine/components/<name>.md`** is created when a component is about to receive concrete attention. Captures interface intent, internal model decisions, open questions, and constraints. Dissolved into source when the component reaches a working MVP; the dissolution is recorded as a reset transaction.

### Game design documents

**`design/game/premise.md`** stays short and aspirational. A page or two. Not a specification. Its job is to provide design pressure for engine validation and a coherent target for game-side work.

Additional game design documents are created as the reference game's design surfaces concrete needs. They follow the same documentation-decay discipline as engine docs.

When a design doc section becomes describable from code alone, remove it. The reset log notes the absorption.

## Decisions Log

`history/decisions.md` is append-only. Each entry captures a significant architectural decision at the moment it is made.

Entry shape:

```
## <date> - <short title>

Scope: engine | game | project
Context: what situation prompted the decision
Decision: what was chosen
Reasoning: why this option over the alternatives considered
Alternatives: what else was considered and why rejected
```

The `Scope` field makes it visible at a glance which side of the engine/game line a decision applies to. `project` covers cross-cutting concerns like workflow, tooling, or repository structure.

**Append-only discipline.** Entries are never edited in place. When a decision is reversed or superseded, a new entry is appended that references and supersedes the prior one. The original stays as a historical record. This keeps the log trustworthy as a record of how thinking evolved.

Decisions worth logging include but are not limited to: choice of dependency, inner-tier architectural commitments, file format and protocol decisions, threading and lifecycle models, and any decision the future self will want to reconstruct the reasoning for.

## Reset Protocol

A reset is a bookkeeping transaction that brings design documentation and code back into alignment. It is event-triggered, not time-triggered.

**Reset triggers:**
- A component crosses from design into working code.
- An inner-tier module stabilizes.
- Visible drift between design docs and current code.
- A session reveals implicit conflicts between accumulated decisions.

**Reset operation.** Every piece of forward-looking design context faces a binary choice:

- **Integrated** - absorbed into code. The design text is removed. The reset summary records what was absorbed and where.
- **Retained** - still forward-looking. Stays in place.

There is no third "completed but still documented" state. That state is the drift this protocol exists to prevent.

**Reset transaction summary.** Appended to `history/resets.md` for each reset. Captures:

```
## <date> - <short title>

Scope: engine | game | project
Trigger: what prompted the reset
Integrated: design context absorbed into code, with pointers to source
Culled: design context removed as obsolete or superseded
Retained: forward-looking context that remains
Decisions promoted: any decisions written to the decisions log during this reset
```

Summaries are short. A reset is bookkeeping, not narrative.

## When This Skill Activates

Whenever work is happening in this repository. The discipline applies to:

- Adding to or modifying any file under `design/`
- Recording entries in `history/decisions.md` or `history/resets.md`
- Evaluating a new dependency
- Deciding whether a piece of design context has been resolved into code
- Making architectural decisions about engine, components, or game systems
- Initiating a context reset
- Deciding whether a surfaced requirement belongs in engine or game code

Before any of the above, this skill is consulted to ensure the working principles, conventions, and protocols are applied.

## Initial State

On project initialization, this file exists alone. The first concrete work is to draft `design/engine/runtime.md` as the load-bearing architectural document, then `design/engine/components.md` as a shallow inventory. `design/game/premise.md` is drafted in parallel or shortly after, with the explicit understanding that it informs engine validation rather than engine architecture. `history/decisions.md` and `history/resets.md` are created empty and populated as events warrant.

No code exists yet. The first code is written when the design documentation is sufficient to make the next concrete step obvious - and not before that point and not after it.
