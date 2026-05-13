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
~/code/curiosity/                       planning workspace (this git repo)
  .claude/skills/curiosity/             project skill
    SKILL.md                            this file
    design/                             codified, validated intent
      engine/
        runtime.md                      core engine runtime layout
        components.md                   index of inner/outer component specs
        components/<name>.md            per-component depth, created on demand
      game/
        premise.md                      aspirational target for the reference game
        <other game design docs>        created as game design surfaces them
    concepts/                           unvalidated candidates under consideration
      engine/<topic>.md                 candidate engine concepts
      game/<topic>.md                   candidate game concepts
    history/
      decisions/                        append-only architectural decisions
        D-###.<slug>.md                 one file per decision
      resets/                           append-only reset transaction summaries
        R-###.<slug>.md                 one file per reset
    resources/
      assets.md                         curated asset acquisition resources
    code/
      conventions.md                    Go conventions for engine and game code
      templates/                        scaffolding templates
  experiments/                          hands-on R&D, tracked in this repo
    <experiment-name>/
      README.md                         Question / Approach / Finding
      <go source>                       prototype code (exempt from conventions)
  <engine, game, ...>                   source repos, gitignored from this workspace
```

Source layout reflects the engine/game separation: engine code and game code live in distinct trees as gitignored sibling directories under the workspace, each with its own git history. The engine has no compile-time dependency on game code.

The `design/` and `concepts/` directories are paired and mirror each other's scope tree (`engine/`, `game/`, project-level). `design/` holds codified intent; `concepts/` holds unvalidated candidates. Concepts are promoted to design only via deliberate bookkeeping in the reset log.

The `experiments/` directory holds standalone Go programs for hands-on validation of concepts that paper alone cannot settle. Experiments are tracked in this repo because their findings inform the planning surface; they are not separate projects, and their code is short-lived (an experiment is removed once its job is done — see the Reset Protocol section).

The `resources/` directory holds reference material that is not subject to the documentation-decay discipline — curated lists, acquisition references, external pointers. It exists alongside the planning surface but follows different rules: entries are added or revised as the external landscape shifts, not as code absorbs them.

## Design Documentation Conventions

Design documents are reference material, not historical record. They describe the present forward-looking state of intent for things not yet built or not yet fully expressed in code. The planning surface separates two claim qualities by directory:

- **`design/`** — codified, validated intent. Claims grounded in decision-log entries, source code, or hard external constraints.
- **`concepts/`** — unvalidated candidates under consideration. Ideas that are not yet ready for design, including open questions and exploratory directions.

A reader can identify a claim's status by the file's directory location, without parsing the prose for confidence cues.

### Design vs. concepts

The two directories mirror each other's scope tree (`engine/`, `game/`, project-level), so a concept's eventual design home is structurally predictable. A concept is promoted to design only via deliberate bookkeeping in the reset log; promotion requires that the writer can articulate why the concept now meets the tangible-and-realistic bar — typically because a decision-log entry was appended, source code was written, or an experiment produced a concrete finding.

Concepts that are falsified, superseded, or no longer load-bearing are culled, not edited indefinitely toward viability. Culling is also bookkeeping in the reset log.

### Engine design documents

**`design/engine/runtime.md`** is the load-bearing engine document. It captures runtime characteristics holistically — the inner-tier boundaries (the systems that share memory and frame lifecycle: ECS, voxel data, physics, rendering), the outer-tier contract (the interface plug-in components implement), and the runtime's own responsibilities (frame loop, scheduling, lifecycle, resource ownership). Per-sub-system detail lives in `components.md` and the per-component files it indexes. Referenced by everything else and the most expensive to get wrong.

**`design/engine/components.md`** is the index of inner and outer component specifications. Each entry is one to two lines — the component's name, a brief responsibility, and a link to its per-component depth file (`design/engine/components/<name>.md`) when that file exists. The index sits empty until the first per-component depth file is created. It does not duplicate runtime.md's tier-placement framing, and it does not carry interface shape (interface shape is concept-tier until firm, per D-010).

**`design/engine/components/<name>.md`** is created when a component is about to receive concrete attention. Captures interface intent, internal model decisions, and constraints not yet enforced in code. Material that remains unsettled lives under `concepts/engine/`, not here. Dissolved into source when the component reaches a working MVP; the dissolution is recorded as a reset transaction.

Engine concepts that are not yet ready for design live under `concepts/engine/` with the same `<topic>.md` naming convention.

### Game design documents

**`design/game/premise.md`** stays short and aspirational. A page or two. Not a specification. Its job is to provide design pressure for engine validation and a coherent target for game-side work.

Additional game design documents are created as the reference game's design surfaces concrete needs. They follow the same documentation-decay discipline as engine docs.

Game concepts that are not yet ready for design live under `concepts/game/`.

When a design doc section becomes describable from code alone, remove it. The reset log notes the absorption.

### Experiments

Some questions cannot be settled on paper. When validation requires hands-on exploration, an experiment is created under `experiments/<name>/`. Each experiment carries a `README.md` with three sections: *Question* (what the experiment is trying to determine), *Approach* (how it goes about answering), and *Finding* (what the experiment showed). The Finding section may be empty while the experiment is in flight; an empty Finding marks the experiment as unfinished.

Experiments are exempt from `code/conventions.md`. They are allowed to be quick and dirty — no `doc.go`, no unit-test ceremony, no convention enforcement.

Completed experiments are retained as historical artifacts (D-033). The Finding section captures the experiment's durable conclusion; the directory persists past the milestone that informed engine or game source. **Success path** — experiment validates a concept; concept is promoted to design; design is integrated into source code. The Finding is updated as each milestone lands. **Failure path** — experiment falsifies or refines a concept; the finding is recorded; the affected concept is updated or culled. The experiment stays as the source record of the investigation. The isolation under `experiments/` keeps historical material structurally separated from production engine and game code.

**No-graduation rule.** Experiment code never becomes engine or game code. Findings inform fresh implementation written against the validated design; the prototype is not lifted forward. The act of writing fresh implementation against a design is itself part of the validation.

## Decisions Log

`history/decisions/` is append-only. Each decision is one file named `D-###.<slug>.md`. Numeric prefix preserves chronological scan order in `ls`; slug provides discoverability via filename. No index file — the directory listing is the index.

Entry shape:

```
# D-### — <short title> — <date>

**Scope**: engine | game | project

**Context**: what situation prompted the decision

**Decision**: what was chosen

**Reasoning**: why this option over the alternatives considered

**Alternatives**: what else was considered and why rejected
```

The `Scope` field makes it visible at a glance which side of the engine/game line a decision applies to. `project` covers cross-cutting concerns like workflow, tooling, or repository structure.

**Append-only discipline.** Files are never edited in place. When a decision is reversed or superseded, a new file is added that references and supersedes the prior one. The original stays as a historical record. This keeps the log trustworthy as a record of how thinking evolved.

**Compaction exception.** A *decision compaction pass* (see Compaction Operations) is the only sanctioned operation that absorbs, compacts, discards, or removes entries. The pass is itself a recorded transaction.

Decisions worth logging include but are not limited to: choice of dependency, inner-tier architectural commitments, file format and protocol decisions, threading and lifecycle models, and any decision the future self will want to reconstruct the reasoning for.

## Reset Protocol

A reset is a bookkeeping transaction that aligns design documentation, concepts, experiments, and code. Every session's closeout includes a reset (the post-session planning discussion in `.claude/behavior/execution.md` produces it). Additional resets fire event-triggered when drift is observed outside a session boundary.

**Standing reset trigger:**
- The post-session planning discussion at every session's closeout.

**Event-triggered (outside the session boundary):**
- A component crosses from design into working code.
- An inner-tier module stabilizes.
- Visible drift between design docs and current code.
- A session reveals implicit conflicts between accumulated decisions.

**Reset operation.** Every piece of forward-looking context faces a choice:

- **Integrated** — absorbed into code or implementation. The design or concept text is removed; the reset summary records what was absorbed and where. Findings from completed experiments are integrated when they reshape a concept or feed a design.
- **Promoted** — concept is promoted to design. The concept file moves into the design tree; the reset summary records the originating concept and its new design home.
- **Culled** — context removed as obsolete, superseded, or falsified. Applies to design sections and concepts. Experiments are not culled — once complete, they remain as historical artifacts (D-033).
- **Retained** — still forward-looking. Stays in place.

There is no "completed but still documented" state. That state is the drift this protocol exists to prevent.

**Reset transaction summary.** Each reset is one file in `history/resets/` named `R-###.<slug>.md`. Same naming and append-only discipline as the decisions log. Captures:

```
# R-### — <short title> — <date>

**Scope**: engine | game | project

**Trigger**: what prompted the reset

**Integrated**: context absorbed into code or implementation, with pointers to source

**Promoted**: concepts that became design entries, with traceability to originating concept

**Culled**: context removed as obsolete, superseded, or falsified

**Retained**: forward-looking context that remains

**Decisions promoted**: decisions written to the decisions log during this reset

**Next session focus**: concrete description of the next session's target and type
                        (development | context | experiment)
```

Summaries are short. A reset is bookkeeping, not narrative.

The `Next session focus` field is the artifact that connects sessions — written at one session's close, read at the next session's start as the orientation seed.

**Cull exception.** A *reset cull pass* (see Compaction Operations) is the only sanctioned operation that removes resets or renumbers them.

## Compaction Operations

Compaction operations are user-invoked maintenance passes that reshape the history logs against current reality. They are recorded exceptions to the append-only discipline of the decisions and resets logs — invoked on demand when the logs feel heavy, never as part of normal session flow.

Two passes are defined, each invoked independently:

- **Decision compaction pass** — reshapes `history/decisions/` against the current state of the skill, behavior, conventions, and source.
- **Reset cull pass** — removes resets whose transactions are no longer forward-looking and renumbers survivors.

### Decision compaction pass

For each file in `history/decisions/`, triage into one of four buckets:

- **Discard** — superseded by a later decision with clear evolution of thought. Discard the superseded entry; the later decision stands as the live one. If the contradiction is not a clean evolution, pause and align with the user before acting.
- **Absorb** — substance reads as a durable convention or principle. Write the convention as a present-tense rule into its natural home: `SKILL.md` (working agreement), a `.claude/behavior/<file>.md` (operational behavior), `code/conventions.md` (source-code style), or source code itself when the convention is structurally enforced. Revise existing partial mentions; never duplicate. Remove the standalone D-### file after absorption.
- **Compact** — substance is a point-in-time architectural or implementation choice now embodied in code but worth preserving for "why was it built this way." Append to `history/decisions/archive.md`, preserving the original D-### prefix inline. Remove the standalone file.
- **Retain** — substance is still load-bearing as a live, standalone decision (recent, hotly-contested, or actively cited by ID elsewhere). Leave the file unchanged.

The pass produces a reset entry with `Scope: project` and `Trigger: Decision compaction pass`, using the standard reset transaction shape. The body lists each decision under one of `Absorbed:`, `Compacted:`, `Discarded:`, `Retained:` — absorptions name the destination file and section; discards name the superseding D-###. The reset entry is the audit trail.

**`archive.md` shape.** Single rolling file at `history/decisions/archive.md`. Header describes its role. Entries appended in original chronological order. Each entry preserves the original `D-###` prefix, title, date, and scope, and collapses the original Context / Decision / Reasoning / Alternatives fields into a single paragraph. Original verbose entries remain recoverable from git history.

### Reset cull pass

For each file in `history/resets/`, evaluated during planning against what the upcoming work will need:

- **Still relevant?** A reset is retained if its Retained, Promoted, or Next-session-focus content still informs work not yet completed. Typical signals: a named concept remains in `concepts/`, a promotion has not yet been integrated into source, or the next session's focus has not yet been honored.
- **Otherwise cull.** Remove the file. Resets are highly volatile and not referenced by R-### outside `history/`, so removal is safe.

Recency is not by itself a retention criterion. The most recent reset is usually retained because its Next-session-focus is typically still load-bearing, but that is an emergent property of the relevance test, not an axiom.

After culls, renumber survivors starting from R-001 in original chronological order. Add a one-line maintenance note at the top of the now-renumbered most recent reset's body: `Maintenance: a Reset Cull Pass was run on YYYY-MM-DD; N resets were culled and survivors renumbered.` No separate audit file — git history is the deeper trail.

## When This Skill Activates

Whenever work is happening in this repository. The discipline applies to:

- Adding to or modifying any file under `design/` or `concepts/`
- Promoting a concept to design, or culling a concept
- Creating, advancing, or removing an experiment under `experiments/`
- Recording entries in `history/decisions/` or `history/resets/`
- Evaluating a new dependency
- Deciding whether a piece of design context has been resolved into code
- Making architectural decisions about engine, components, or game systems
- Initiating a context reset
- Deciding whether a surfaced requirement belongs in engine or game code

Before any of the above, this skill is consulted to ensure the working principles, conventions, and protocols are applied.
