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

**Joint evaluation.** Architectural alternatives are judged across performance, implementation complexity, and development ergonomics together — not by performance alone. Complexity carries cost over the project's lifetime; ergonomics determines what the user can confidently reason about and modify. The winner is the alternative with the highest joint score, not the highest performance score in isolation. Performance gaps must be substantial and load-bearing to justify substantial complexity or ergonomic cost. The judgment is qualitative; there are no formal weights or scoring rubrics.

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
  engine/                               in-tree during prototype phase
  go.work                               unifies engine + any in-tree modules
  <game, outer-tier modules>            separate repos, gitignored from this workspace
```

Source layout reflects the engine/game separation. During the prototype phase, the engine module lives in-tree at `~/code/curiosity/engine/` as a self-contained Go sub-module of the curiosity repository; once the engine reaches release stability, it graduates to its own repository with separate git history. Game and outer-tier-implementation modules remain gitignored sibling directories under the workspace, each carrying its own git history. The engine has no compile-time dependency on game code.

The `design/` and `concepts/` directories are paired and mirror each other's scope tree (`engine/`, `game/`, project-level). `design/` holds codified intent; `concepts/` holds unvalidated candidates. Concepts are promoted to design only via deliberate bookkeeping in the reset log.

The `experiments/` directory holds standalone Go programs for hands-on validation of concepts that paper alone cannot settle. Experiments are tracked in this repo because their findings inform the planning surface; they are not separate projects, and their code is short-lived (an experiment is removed once its job is done — see the Reset Protocol section).

The `resources/` directory holds reference material that is not subject to the documentation-decay discipline — curated lists, acquisition references, external pointers. It exists alongside the planning surface but follows different rules: entries are added or revised as the external landscape shifts, not as code absorbs them.

## Context surfaces

The official-context surface — every file outside `history/` — is the project's durable reference. Each surface carries a topic, a volatility profile, and a belongs-here test. The compaction protocol consults this section to find homes for substance graduating out of volatile context, and everyday writing consults it to choose between sibling destinations when a piece of substance could plausibly live in more than one place.

`history/` is the boundary case: deliberately volatile context that churns through sessions and compaction passes. Substance there is either graduating out — to one of the surfaces below — or being culled. The discipline of this section is what makes graduation mechanical instead of judgment-grade.

Surfaces are grouped below by cluster. Each entry names what the surface represents, how stable its content is, how its content is organized internally, and the test that decides whether new substance belongs here versus a sibling.

### Workspace conduct

**`.claude/CLAUDE.md`** is the workspace-level master directive file, loaded automatically every session. Stable and short: session-start posture, partnership disciplines, the behavior-layer index. Content is organized as terse master sections that point downward to the behavior layer for detail. *Belongs here*: directives that govern every session regardless of phase or task. *Doesn't*: per-phase operational detail (→ `.claude/behavior/`), project workflow (→ `SKILL.md`), Go style (→ `code/conventions.md`).

**`.claude/behavior/<name>.md`** is the topical expansion of workspace conduct, one file per cluster (`execution`, `source-code`, `communication`, `collaboration`, `verification`). Loaded on demand per CLAUDE.md's *Load when* descriptions. Moderately stable — revised as the collaboration model evolves, not on every session. Each file stands on its own when loaded in isolation; sections within a file are themselves topical. *Belongs here*: operational rules for one cluster of work, with the depth to apply the cluster without cross-loading. *Doesn't*: terse master directives (→ `CLAUDE.md`), project-level discipline (→ `SKILL.md`), Go style (→ `code/conventions.md`).

### Project skill

**`.claude/skills/curiosity/SKILL.md`** is the project's working agreement: scope, principles, dependency policy, repository layout, this inventory, design-documentation conventions, history-log shapes, and protocols (reset, compaction). Loaded whenever the curiosity skill is invoked. Stable — revisions are deliberate and recorded as compaction-pass entries or session-closeout updates. Content is organized as topical sections; each section is the canonical home for its concern. *Belongs here*: project-wide rules, principles, and protocols that apply across engine and game work. *Doesn't*: workspace conduct (→ `CLAUDE.md` + `.claude/behavior/`), Go source-code style (→ `code/conventions.md`), forward-looking engine architecture (→ `design/engine/` or `concepts/engine/`).

**`.claude/skills/curiosity/code/conventions.md`** is the Go source-code style reference for engine and game code, distilled from the tau and herald codebases. Stable reference material — consulted at module-creation time and when an unsettled question surfaces. Not subject to documentation-decay discipline; entries codify patterns rather than describe forward-looking work. Organized as numbered sections (workspace topology, package documentation, interfaces, constructors, error model, testing, concurrency, etc.). *Belongs here*: Go style rules, layout conventions, idiom selections that apply across packages and persist through API churn. *Doesn't*: project workflow (→ `SKILL.md`), per-package design (→ `design/engine/`), workspace conduct (→ `CLAUDE.md` + behavior/).

**`.claude/skills/curiosity/code/templates/`** holds scaffolding templates (`doc.go.tmpl`, `CHANGELOG.md.tmpl`) referenced by `conventions.md`. Stable — revised when convention changes mandate template updates. Each template is its own file. *Belongs here*: literal scaffolding to be instantiated at package- or module-creation time. *Doesn't*: prose conventions about when to use the templates (→ `conventions.md`).

### Forward-looking planning

**`.claude/skills/curiosity/design/engine/`** and **`design/game/`** hold codified, validated intent for things not yet built or not yet fully expressed in code. Settled — claims are grounded in source code, hard external constraints, or accumulated agreement among prior decisions and concept work. Revisable when constraints shift, but revision is deliberate. Engine design is organized hierarchically (`runtime.md` for the load-bearing engine document, `components.md` as the per-component index, `components/<name>.md` for depth on individual components). Game design is shallower, anchored by `premise.md`. Both follow the documentation-decay discipline: sections collapse when code can express them. *Belongs here*: forward-looking intent for engine or game architecture that is settled enough that the reader can rely on it. *Doesn't*: open questions or candidate approaches not yet decided (→ `concepts/`), historical reasoning about past decisions (lives in surrounding text or git history), source-code style (→ `code/conventions.md`).

**`.claude/skills/curiosity/concepts/engine/`** and **`concepts/game/`** hold unvalidated candidates under consideration — open questions, exploratory directions, intent that is not yet ready for design. Deliberately malleable. The two trees mirror `design/`'s scope structure so promotion is structurally predictable. Game concepts default to staying malleable longer than engine concepts; engine concepts have more anchors in physical and computational constraint. Each file is one topical question (`<topic>.md`) with sections like *Question*, *Constraints already locked in*, *Open design surface*. Concepts are promoted to `design/` only via deliberate bookkeeping in the reset log; concepts that are falsified, superseded, or no longer load-bearing are culled. *Belongs here*: questions actively in play whose answers will eventually shape engine or game design but are not yet settled. *Doesn't*: settled intent (→ `design/`), historical "we picked A over B" narratives (lives in surrounding text or git history).

### External pointers

**`.claude/skills/curiosity/resources/`** holds reference material whose volatility is external to the project: curated lists, acquisition references, links into ecosystems outside the workspace (e.g., `assets.md`). Updated as the external landscape shifts, not as code absorbs them — explicitly outside the documentation-decay discipline. Each file is one topical resource catalog. *Belongs here*: pointers and curated lists whose value is "where to find X" rather than "what X is." *Doesn't*: working agreements (→ `SKILL.md`), conventions (→ `code/conventions.md`), design intent (→ `design/`).

### Records of investigation

**`experiments/<name>/`** holds hands-on R&D — standalone Go programs that answer questions paper alone cannot settle, plus a `README.md` carrying *Question / Approach / Finding*. Retained as historical artifacts: the directory persists past the milestone it informed, with the Finding section capturing the durable conclusion. Conventions exemption: experiment code is allowed to be quick and dirty (no `doc.go`, no unit-test ceremony). No-graduation rule: experiment code never becomes engine or game code; findings inform fresh implementation written against the validated design. Each experiment is isolated from engine and game source — it cannot drift into production code, only inform it. *Belongs here*: a self-contained investigation with a stated question and a recorded finding. *Doesn't*: production code (→ `engine/`, `game/`, outer-tier siblings), conventions or principles surfaced by the experiment (those graduate to `conventions.md`, `SKILL.md`, or `design/` per the relevant home).

### Source code

**`engine/`**, **`game/`**, and **sibling outer-tier modules** under `~/code/curiosity/` are the project's production code. Volatility varies by package maturity — primitives stabilize early, integration layers churn longer. Engine code is tracked in the curiosity repository during the prototype phase; game and outer-tier modules are gitignored from the workspace and each carry their own git history when they land. Organized per `code/conventions.md` (multi-module workspace, `doc.go` per package once stable, colocated black-box tests, etc.). *Belongs here*: implementation — types, functions, methods, configuration loading, the actual behavior. *Doesn't*: rules about how code is written (→ `code/conventions.md`), forward-looking architecture that hasn't been built (→ `design/` or `concepts/`), experiment prototypes (→ `experiments/`).

### Boundary: volatile context

**`.claude/skills/curiosity/history/decisions/`** and **`history/resets/`** are the volatile context this section frames against. Decisions accumulate as `D-###.<slug>.md` files; resets accumulate as `R-###.<slug>.md` files. Both are subject to compaction operations: decisions graduate to the official-context surfaces above (or are culled); resets are culled when no longer informing upcoming work. *Belongs here*: in-flight architectural reasoning awaiting graduation, and session-bookkeeping transactions. *Doesn't*: anything that has earned a home in official context — that lives in its home, with `history/` carrying at most the audit-trail reset entry that records the graduation.

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

**`design/engine/components.md`** is the index of inner and outer component specifications. Each entry is one to two lines — the component's name, a brief responsibility, and a link to its per-component depth file (`design/engine/components/<name>.md`) when that file exists. The index sits empty until the first per-component depth file is created. It does not duplicate runtime.md's tier-placement framing, and it does not carry interface shape (interface shape is concept-tier until firm).

**`design/engine/components/<name>.md`** is created when a component is about to receive concrete attention. Captures interface intent, internal model decisions, and constraints not yet enforced in code. Material that remains unsettled lives under `concepts/engine/`, not here. Dissolved into source when the component reaches a working MVP; the dissolution is recorded as a reset transaction.

Engine concepts that are not yet ready for design live under `concepts/engine/` with the same `<topic>.md` naming convention.

### Game design documents

**`design/game/premise.md`** stays short and aspirational. A page or two. Not a specification. Its job is to provide design pressure for engine validation and a coherent target for game-side work.

Additional game design documents are created as the reference game's design surfaces concrete needs. They follow the same documentation-decay discipline as engine docs.

Game concepts that are not yet ready for design live under `concepts/game/`.

**Game-side malleability.** Game-side concepts and design hold material malleable longer than the workspace's iterative-depth principle alone would imply. Engine architecture has more anchors in physical and computational constraint, so deepening early is safer there; game design lives downstream of "what feels good in play," which paper alone cannot decide. When tempted to commit to a mechanism, prefer expanding the *Question* in `concepts/game/<topic>.md` — more sub-questions, more constraints made explicit — over proposing the mechanism; let playable validation decide which combinations land. When tempted to add a game-side aspiration that commits the engine to a specific feel, check whether the commitment is engine-pressure (it earns the bullet) or game-design preference (it lives in concepts).

When a design doc section becomes describable from code alone, remove it. The reset log notes the absorption.

### Experiments

Some questions cannot be settled on paper. When validation requires hands-on exploration, an experiment is created under `experiments/<name>/`. Each experiment carries a `README.md` with three sections: *Question* (what the experiment is trying to determine), *Approach* (how it goes about answering), and *Finding* (what the experiment showed). The Finding section may be empty while the experiment is in flight; an empty Finding marks the experiment as unfinished.

Experiments are exempt from `code/conventions.md`. They are allowed to be quick and dirty — no `doc.go`, no unit-test ceremony, no convention enforcement.

Completed experiments are retained as historical artifacts. The Finding section captures the experiment's durable conclusion; the directory persists past the milestone that informed engine or game source. **Success path** — experiment validates a concept; concept is promoted to design; design is integrated into source code. The Finding is updated as each milestone lands. **Failure path** — experiment falsifies or refines a concept; the finding is recorded; the affected concept is updated or culled. The experiment stays as the source record of the investigation. The isolation under `experiments/` keeps historical material structurally separated from production engine and game code.

**No-graduation rule.** Experiment code never becomes engine or game code. Findings inform fresh implementation written against the validated design; the prototype is not lifted forward. The act of writing fresh implementation against a design is itself part of the validation.

**Measurement integrity.** When an experiment's job is to measure cost — throughput, latency, allocation, memory, whatever the comparison depends on — the implementation faithfully reproduces the cost profile of a production-grade version of what is being measured, even when the chosen workload does not exercise that cost at runtime. A simplification that compresses measured cost relative to production is measurement bias by omission, and is not acceptable regardless of whether it is documented. The conventions exemption above governs style and ceremony; it does not govern measurement validity. Honesty in measurement is what the experiment produces; everything else (code reuse, brevity, exemption from style rules) is subordinate to it.

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

**Compaction exception.** A *decision compaction pass* (see Compaction Operations) is the only sanctioned operation that graduates or culls entries. The pass is itself a recorded transaction.

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
- **Culled** — context removed as obsolete, superseded, or falsified. Applies to design sections and concepts. Experiments are not culled — once complete, they remain as historical artifacts.
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

The pass triages every file in `history/decisions/` into one of two outcomes — **Graduate** or **Cull**. There is no archive: a decision either earns a home in official context or it is removed. Substance worth preserving graduates; substance not worth preserving is culled. Retention as a standalone D-### file between passes is the normal state and is not a pass outcome.

**Mental model.** `history/` is volatile context that churns through sessions. Everything outside `history/` is official context, organized hierarchically by topic per *Context surfaces*. Compaction is the process by which volatile substance promotes into official context — or is dropped because it does not earn a place there. The discipline that protects this is the proliferation guard: do not manufacture a home for a decision that does not have one. The hierarchy is meant to stay organized, tidy, and easy to follow.

**Graduate.** A decision graduates when its substance is load-bearing for future work AND a topically natural home exists in official context. The substance is written into that home as a present-tense rule, fact, or design statement; the standalone D-### file is removed; all `(D-###)` citations to it elsewhere are stripped or re-referenced so the surviving text reads without the decision pointer. *Context surfaces* names where each kind of substance belongs: working agreement → `SKILL.md`; workspace conduct → `CLAUDE.md` or `.claude/behavior/`; Go style → `code/conventions.md`; settled engine intent → `design/engine/`; open engine questions → `concepts/engine/`; source-level enforcement → engine or game source.

**Cull.** A decision is culled when (a) it is contradicted by a superseding decision such that it should not survive into official context, or (b) its substance is not load-bearing enough to warrant a home in the topical hierarchy. The standalone file is removed; no substance is preserved on a sidecar surface. Forward-looking citations to an unbuilt decision (e.g., a concept that cites "(D-###)" as the source of a constraint) are a signal that the decision belongs in the cited concept itself — not a reason to keep it standalone.

**Citation discipline.** `(D-###)` citations belong only inside `history/`. After a decision graduates, its substance is the surrounding text in its destination; the citation pointer is redundant. The pass strips citations across the official-context surface so the prose reads on its own terms. Where a citation is paired with a destination pointer (e.g., `(D-002; design/engine/runtime.md — Inner-tier members)`), the `(D-###)` prefix is stripped and the pointer is kept.

**Pause points.** Pause and align with the user when:
- The contradiction with a superseder is not a clean evolution.
- A Graduate candidate's natural home is ambiguous across multiple topical surfaces.
- A decision feels Cull-eligible but the load-bearing call is judgment-grade.

**Audit trail.** The pass produces a reset entry with `Scope: project` and `Trigger: Decision compaction pass`, using the standard reset transaction shape. The body lists each decision under `Graduated:` (with destination file and section) or `Culled:` (with rationale: contradicted-by or no-home). The reset entry plus git history is the audit trail; there is no separate archive.

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
