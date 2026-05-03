# Decisions Log

Append-only record of significant architectural decisions for the curiosity project. Each entry captures the situation, the choice, the reasoning, and the alternatives at the moment of decision. Entries are never edited in place; supersession is achieved by appending a new entry that references and replaces the prior one.

---

## D-001 — Project scope: engine primary, reference game secondary — 2026-05-02

**Scope**: project

**Context**: Initial project scoping in the planning conversation. The user is starting a Go-based voxel game engine project but also has aspirational ideas for a specific game (a post-organic earth premise). Question: is the engine the artifact, or is the game?

**Decision**: The engine is the primary artifact. A reference game is developed alongside it as a forcing function and validation target — but engine decisions serve the generic voxel-game case, not premise-specific needs. Features that only make sense for the reference game live in game-side code.

**Reasoning**: Engines designed in pure isolation from any game tend to optimize for imagined requirements. Engines designed around a single game tend to bake in premise-specific assumptions and limit reuse. The reference-game-as-forcing-function pattern (Herald serving as TAU's proof-of-concept is the precedent) gets the benefit of concrete validation without coupling. The user's stated preference is to focus on engine foundations; this framing matches that intent.

**Alternatives**: (1) Build the engine for the specific game from the start — rejected as it bakes in premise-specific assumptions and limits reuse. (2) Build the engine in isolation with no reference game — rejected as it loses the validation signal that prevents architecturally-correct-but-practically-wrong design.

---

## D-002 — Two-tier architecture: co-designed inner core, pluggable outer libraries — 2026-05-02

**Scope**: engine

**Context**: User initially proposed building all engine subsystems as standalone libraries (voxels, rendering, audio, UI, physics, networking, storage, ECS). Discussion surfaced that some subsystems share hot-path data and cannot be designed independently without significant rework later.

**Decision**: Two-tier architecture. The inner tier (ECS, voxel data, physics, rendering) is co-designed — sharing memory layout, threading model, and frame lifecycle, with internal module boundaries but shared assumptions. The outer tier (audio, UI, storage, networking, content pipeline) is genuinely standalone — independent libraries with stable contracts plugging into the engine.

**Reasoning**: This matches how production engines (Bevy, Unity DOTS, Unreal) are actually organized despite their modular marketing. The inner tier's performance characteristics make boundary-piercing inevitable when interfaces are too pure; designing the inner-tier modules together avoids the "interfaces become the bottleneck" trap. The outer tier's cold-path nature makes clean contracts cheap and modular development genuinely viable.

**Alternatives**: (1) Flat library structure with all subsystems as peers — rejected as it ignores the hot-path coupling between ECS, voxels, physics, and rendering. (2) Monolithic engine with no module boundaries — rejected as it loses the testability and learning value of modular design.

---

## D-003 — Workflow discipline: single source of truth, plan-only-next-step, documentation decay, append-only history, event-triggered resets — 2026-05-02

**Scope**: project

**Context**: User wants a development workflow that prevents documentation drift and supports iterative discovery. The failure mode being prevented is design docs that describe what the code already does, then diverge over time and become misleading.

**Decision**: Adopt five disciplines: (1) Single source of truth — every piece of context lives in exactly one place (code or design doc, never both). (2) Plan only the next step — no roadmaps, no execution ordering, no phases. (3) Documentation decay — design docs only contain what code cannot express; absorbed sections are removed. (4) Append-only history logs — decisions and resets accumulate; entries are never edited in place. (5) Event-triggered resets — bookkeeping happens when components stabilize or drift surfaces, not on a schedule.

**Reasoning**: The single-source-of-truth rule plus documentation decay enforces alignment between docs and code. The append-only history rule preserves reasoning over time without encouraging revisionism. Event triggers match natural project rhythms better than calendar resets, which fire when nothing has changed or miss the moments when everything has.

**Alternatives**: (1) Calendar-based resets (weekly, monthly) — rejected as misaligned with project rhythm. (2) Editable history — rejected as it destroys the record of how thinking evolved. (3) Roadmap-driven planning — rejected as the destination is genuinely unsettled and roadmaps would over-commit.

---

## D-004 — Dependency policy: active+supported, open source, critical-to-execution; standard library preferred; engine and game evaluated independently — 2026-05-02

**Scope**: project

**Context**: User stated learning is as much a goal as building, and wants minimal dependencies. Need a concrete policy to evaluate every "do we take this on?" question consistently.

**Decision**: A dependency is taken on only when all three conditions hold: (1) active and well-supported (recent commits, responsive maintainership, healthy issue activity); (2) open source (no closed or source-available licenses); (3) critical to execution (project cannot reasonably proceed without it, or reimplementation would consume effort better spent on the project's core). Standard library preferred wherever it suffices. Engine and game dependencies evaluated independently — game-side dependencies do not flow upward into the engine.

**Reasoning**: Minimizing dependencies serves both the learning goal and the engineering goal. Each new direct dependency is recorded in the decisions log with the reasoning that satisfied the three conditions, so future review is grounded in the original justification.

**Alternatives**: (1) Whitelist-only approach (hard list of allowed deps) — rejected as too rigid for an exploratory project. (2) No restrictions, evaluate ad-hoc — rejected as it tends to accumulate unnecessary deps over time.

---

## D-005 — Go conventions adopted from tau/herald codebases; inner-tier divergence explicit — 2026-05-02

**Scope**: engine

**Context**: User has mature Go codebases (`~/tau/{protocol,format,provider,agent,orchestrate,examples}` and `~/code/herald`) that exemplify a coherent and tested working style. Need a consistent convention layer for the engine project rather than reinventing patterns from scratch.

**Decision**: Adopt the Go conventions observed across tau and herald, distilled into `code/conventions.md`. Multi-module workspaces; `doc.go` per package (deferred until stability); small consumer-site interfaces; `New(cfg, deps) Type` constructors with ephemeral config; three-phase config lifecycle; sentinel errors plus struct-based custom errors with functional options; black-box testing in sibling `tests/` directory; goroutine-per-message at the outer tier; `sync.RWMutex` registries; strategic generics. The inner tier (ECS, voxel data, physics, rendering) explicitly diverges where hot-path constraints justify.

**Reasoning**: These conventions have held up under load in production code; reusing them avoids reinventing solved problems and keeps the project consistent with the user's existing style. The inner-tier carve-out prevents naive application of cold-path conventions to hot-path code.

**Alternatives**: (1) Generate fresh conventions from scratch — rejected as it discards a known-good style without justification. (2) Apply tau conventions uniformly without inner-tier carve-out — rejected as it ignores known performance traps in voxel-engine hot paths.

---

## D-006 — Workspace topology: superproject at ~/code/curiosity holds skill+history+resources+code; sub-projects live as separate gitignored repos in subdirectories — 2026-05-02

**Scope**: project

**Context**: Need to decide where the engine, game, and tooling repositories live relative to the planning workspace, and how cross-machine sync of the planning surface works.

**Decision**: `~/code/curiosity/` is the planning superproject — a git repo holding only the durable cross-machine context (skill, design, history, resources, code conventions, templates, behavior layer). Sub-projects (engine modules, the reference game) live in subdirectories under the same path but are `.gitignore`d from the workspace repo and maintain their own independent git history.

**Reasoning**: This separation lets the planning surface persist independently of the source code. Each sub-project gets its own license, CHANGELOG, and release cadence without polluting the planning history. Cross-machine sync via the workspace repo does not drag full source history along with it.

**Alternatives**: (1) Monorepo with everything in one repository — rejected as it couples planning state to source state and complicates per-sub-project release workflows. (2) Planning as a separate top-level directory unrelated to source — rejected as it makes navigating between planning and source require directory-changing.

---

## D-007 — Skill scope and name: project skill at .claude/skills/curiosity, named "curiosity" (placeholder) — 2026-05-02

**Scope**: project

**Context**: Need to decide whether the project skill is workspace-scoped or user-scoped, and pick an initial name. The engine, game, and organization names are not yet formalized.

**Decision**: Project skill at `~/code/curiosity/.claude/skills/curiosity/`. Skill name `curiosity` as a placeholder until the engine, game, and organization names are formalized.

**Reasoning**: Workspace-scoped persists across machines via the workspace git repo, which matches the cross-machine sync goal. Project-scoped also activates only in this workspace, avoiding pollution of unrelated work. The `curiosity` placeholder reuses the workspace codename — one name to track during the placeholder phase.

**Alternatives**: (1) User-scoped skill in `~/.claude/skills/` — rejected as it doesn't persist via the workspace repo and activates everywhere. (2) Initial name `voxel-engine-project` — rejected as it adds a second name to manage alongside `curiosity`.

---

## D-008 — GitHub used for source control only; no project-management features — 2026-05-02

**Scope**: project

**Context**: User created `github.com/JaimeStill/curiosity` as the remote and stated they want to avoid GitHub project-management features for this project. The reasoning: keep all context embedded in the repository to preserve single source of truth and avoid context-switching.

**Decision**: GitHub Issues, Projects, Discussions, Milestones, and Wiki are intentionally unused. All planning context (design, decisions, resets, asset references, conventions, behavior layer) lives in the repository itself. GitHub serves only as source control and remote backup.

**Reasoning**: Single source of truth — context never has to be reassembled across surfaces. Avoids context-switching and API calls to external planning artifacts. Each sub-project repo (engine, game) makes its own decision on this when spun up; the workspace policy does not propagate automatically.

**Alternatives**: (1) Use GitHub Issues for tracking — rejected as it splits planning context across surfaces. (2) Use a separate project-management tool (Linear, Notion, etc.) — rejected for the same reason and adds external dependency.

---

## D-009 — Workspace behavior layer: .claude/CLAUDE.md master index plus five .claude/behavior/ files — 2026-05-02

**Scope**: project

**Context**: Project workflow (in `SKILL.md`) and Claude Code's conduct in the workspace are different concerns. The user's pacing preferences, source-code-authoring boundary, partnership posture, communication style, and verification discipline are about how Claude operates in this workspace, not about project workflow.

**Decision**: Workspace `.claude/CLAUDE.md` is a master index; topical detail lives in five `.claude/behavior/` files (`execution.md`, `source-code.md`, `communication.md`, `collaboration.md`, `verification.md`). CLAUDE.md is always loaded; behavior files load on demand per the `Load when` descriptions in the index. Sessions always initialize in plan mode. Topic-branch-per-session with PRs lifted from `resets.md` closeout entries. The user authors source code; Claude produces supporting artifacts (godoc, tests, design, decisions, conventions, templates) — and these supporting artifacts are written or revised exclusively during the closeout phase, with source-code design driving them, never the reverse.

**Reasoning**: Project workflow (`SKILL.md`) and Claude conduct (behavior layer) are different concerns and benefit from separate artifacts. The index-plus-loaded-on-demand layout keeps the always-loaded surface lean while letting each behavior topic stay focused. The behavior layer is workspace-versioned to persist across machines, same as the skill itself.

**Alternatives**: (1) Fold all behavior into `SKILL.md` — rejected as it conflates project workflow with Claude conduct. (2) Use `~/.claude/CLAUDE.md` (global) — rejected as these directives are not yet proven universal; promote individual directives later if they hold up across multiple projects. (3) Single workspace `CLAUDE.md` without the `behavior/` split — rejected as the directives total 70+ lines and benefit from topical organization with on-demand loading.

---

## D-010 — Design surface split: `design/` for codified intent, `concepts/` for unvalidated candidates; promotion via reset bookkeeping — 2026-05-02

**Scope**: project

**Context**: Plan-mode discussion before drafting `design/engine/runtime.md` surfaced that the proposed document structure would mix locked-in claims (tier criterion, members from D-002) with unsettled candidates (open questions about scheduler shape, threading model, ECS storage strategy). Mixed claim-quality content inside a single design document recreates the documentation drift the workflow exists to prevent — a reader cannot distinguish committed intent from speculative material without parsing carefully, and writers tend to promote unsettled material into the locked-in body via osmosis.

**Decision**: Split the design surface into two directories with distinct claim-quality. `design/` carries codified, validated intent — claims grounded in decision-log entries, source code, or hard external constraints. `concepts/` carries unvalidated candidates — ideas under consideration that have not yet met the bar for design. Both directories mirror the same scope tree (`engine/`, `game/`, project-level). A concept is promoted to design only via deliberate bookkeeping in the reset log; promotion requires that the writer can articulate why the concept now meets the tangible-and-realistic bar.

**Reasoning**: The file's directory location communicates the claim's status without requiring the reader to parse carefully. Promotion via reset bookkeeping creates a forcing function — concepts move only when promotion is defensible. Mirroring the scope tree across both directories makes promotion mechanical (move file, update references) and avoids structural asymmetry between the two surfaces.

**Alternatives**: (1) Single design directory with explicit "Open questions" sections inside each doc — rejected because adjacency between locked-in and unsettled material erodes the discipline; writers feel confident and promote unsettled material into the locked-in body without bookkeeping. (2) Flat `concepts/` directory not mirroring `design/` — rejected because concepts almost always pertain to a specific scope and a flat layout creates promotion friction. (3) Status-per-section markers within design files — rejected as too fine-grained to enforce reliably; directory-level discipline is coarser but more honest.

---

## D-011 — Post-session planning discussion as closeout phase; reset-entry shape grows `Promoted` and `Next session focus`; closeout sequence revised — 2026-05-02

**Scope**: project

**Context**: `.claude/behavior/execution.md` defines a session-start orientation phase (read latest reset entry, scope the agenda) but no symmetric session-end phase. R-001's "Follow-up" line informally pointed to the next session's work, but that was ad-hoc rather than a structural part of the protocol. The design/concepts split (D-010) also requires a deliberate gate for concept promotion — without one, concepts will drift into design via writer confidence rather than explicit bookkeeping.

**Decision**: Add a post-session planning discussion as a distinct phase of closeout, conducted in plan mode. The discussion reviews what the session accomplished, evaluates the state of design and concept documentation, and produces (1) any concept→design promotions, (2) any concept culls, (3) any design absorptions where source landed, (4) the next session's focus. Reset entry shape grows two new fields: `Promoted` (concepts that became design entries during the session, with traceability to the originating concept), and `Next session focus` (concrete description of what the next session targets, including session type). Existing `Integrated` / `Culled` / `Retained` / `Decisions promoted` fields remain. Closeout sequence revised to: (1) godoc, (2) unit tests, (3) post-session planning discussion, (4) documentation/context adjustments, (5) decisions log entries, (6) reset entry, (7) commit / push / PR. No-source sessions skip steps 1–2 gracefully — godoc and tests are no-ops when no source code was produced.

**Reasoning**: Bookends the session symmetrically — every session opens with orientation and closes with the seed of the next session's orientation. Routes all promotions and culls through a structured plan-mode discussion at closeout, replacing implicit drift with deliberate bookkeeping. Closes the documentation-decay loop: design absorbed into code is removed during the discussion rather than waiting to be noticed later. The `Next session focus` field is the artifact that connects sessions — written at one session's close, read at the next session's start.

**Alternatives**: (1) Keep next-session direction as a freeform "Follow-up" line in reset entries — rejected as too easily skipped or written carelessly when it's not a structural part of the protocol. (2) Track next-session focus in a separate file — rejected as it scatters the planning surface; the reset entry already consolidates session history and is the natural home. (3) Make the planning discussion optional — rejected because the discipline only works if it runs every session; optionality reintroduces drift.

---

## D-012 — `experiments/` directory at workspace root; exploratory R&D session type; conventions exemption; no-graduation rule — 2026-05-02

**Scope**: project

**Context**: The design/concepts pipeline (D-010) handles ideas that can be evaluated on paper, but some questions in the voxel-engine domain — meshing strategies, ECS storage layouts, camera control feel, lighting models — cannot be settled without hands-on validation. Without a sanctioned location for that work, the temptation is either to over-deliberate on paper and commit prematurely, or to prototype directly in engine code and pollute the architecture with throwaway material.

**Decision**: Add `experiments/` at the workspace root, tracked in the planning repo alongside `design/`, `concepts/`, and `history/`. Each experiment lives in its own subdirectory containing a `README.md` with three sections (*Question*, *Approach*, *Finding*) plus whatever Go source the experiment requires. `code/conventions.md` does not apply to experiments — they are allowed to be quick and dirty; no `doc.go`, no unit-test ceremony, no convention enforcement. **Lifecycle**: an experiment exists only while its job is active. *Success path* — experiment validates a concept, concept is promoted to design, design is integrated into source code, then the experiment is removed in that integration session's closeout. *Failure path* — experiment falsifies or refines a concept, finding is captured in `concepts/` (concept evolves or is culled), then the experiment is removed in the same session's closeout. **No-graduation rule**: experiment code never becomes engine or game code. Findings inform fresh implementation written against the validated design; the prototype is not lifted forward. **Bookkeeping**: experiment lifecycle events fold into existing reset categories — `Integrated` for findings absorbed, `Culled` for removals — with explicit naming of the concept or design the experiment served, preserving traceability without taxonomy bloat. **Session types**: exploratory R&D becomes a distinct session type alongside development and context sessions; the post-session planning discussion (D-011) determines next session's type as well as its focus.

**Reasoning**: Experiments are planning-tier artifacts that happen to involve code, not engine code in waiting. Tracking them in the planning repo keeps findings adjacent to the concept/design surface they inform. Time-bounded existence prevents experiments from accumulating as ambient evidence that decays alongside the questions they originally answered. The no-graduation rule prevents prototype-grade decisions from smuggling into production-grade code under the cover of "we already wrote it." Exempting experiments from `code/conventions.md` keeps activation cost low — a `doc.go` and unit-test ceremony for a 200-line spike defeats the purpose. Folding lifecycle events into existing reset categories avoids parallel taxonomies for what is fundamentally the same bookkeeping operation.

**Alternatives**: (1) Experiments in a separate gitignored repo — rejected because experiments would not be referenced unless their result is codified into a design, and the deliberate transcription back into the planning surface is itself the forcing function for "what did this experiment actually tell us?". (2) Permanent retention of experiments — rejected because experiments capture findings, not implementations; once the finding is absorbed, the experiment is dead weight whose presence implies it might still be informative. (3) Allow experiment code to be lifted into engine or game source — rejected because the act of writing fresh implementation against a validated design is itself part of the validation; lifting prototype code skips that step. (4) Add a parallel reset taxonomy for experiment lifecycle events — rejected as taxonomy bloat for what cleanly fits as `Integrated`/`Culled` with traceability fields.

---

## D-013 — Role separation between `design/engine/runtime.md` and `design/engine/components.md`: holistic vs. index — 2026-05-02

**Scope**: project

**Context**: Drafting `design/engine/runtime.md` populated the inner-tier and outer-tier member rosters with 2–3 sentences each on responsibility and tier-placement reasoning. SKILL.md's `components.md` description ("shallow by default: a list of inner and outer components with one to two sentences each on responsibility and rough interface shape") then read either as redundant with runtime.md or as describing a different kind of artifact whose role was unclear. The split of responsibility between the two docs needed naming.

**Decision**: `runtime.md` captures runtime characteristics holistically — tier criterion, member rosters, runtime roles — and serves as the engine's load-bearing reference doc. `components.md` is the index of inner and outer component specifications, with each entry pointing to a per-component depth file (`components/<name>.md`) when one exists; the index sits empty until the first depth file is created. Interface shape is not carried in `components.md` because interface shape is concept-tier until firm (per D-010). SKILL.md is adjusted to reflect this division.

**Reasoning**: Two distinct purposes — architectural anchor vs. mechanical inventory — are best served by two distinct docs with named roles. Naming the linkage explicitly prevents content from drifting into the wrong doc. Keeping `components.md` empty until per-component depth exists prevents premature stub material that violates the design/concepts discipline.

**Alternatives**: (1) Cull `components.md` from SKILL.md entirely on the grounds that runtime.md's member sections subsume it — rejected because runtime.md captures tier placement, not per-sub-system detail, and a separate index makes per-component detail discoverable as it accumulates. (2) Allow `components.md` to carry interface shape directly — rejected because interface shape is concept-tier until firm; SKILL.md must respect the design/concepts split.

---

## D-014 — Game-side concepts default malleable longer than the iterative-depth principle alone implies — 2026-05-02

**Scope**: project

**Context**: During R-004 (the design-game-premise session), after drafting `design/game/premise.md` and four companion concept stubs, the user surfaced that game-side design specifically should remain open to optimization through iteration. Their explicit framing: *"There's a lot for us to explore in this space, and I don't want us to get too locked into a particular element or vision. We should optimize what works best and provides the best feeling experiences possible."* Earlier in the same session, about the conflict-and-encounters stub: *"I just want it to be there for us to explore."* The default iterative-depth principle (SKILL.md, "Iterative depth") treats engine and game alike; this guidance asymmetrically prioritizes malleability for game-side material.

**Decision**: For game-side concepts and design (`concepts/game/`, `design/game/`), default to holding material malleable longer than the iterative-depth principle alone implies. Capture intent as questions in `concepts/game/<topic>.md` stubs rather than committing to mechanism specifics; favor letting playable validation decide which combinations feel best. When a depth pass becomes imminent, prefer expanding the *Question* (more sub-questions, more constraints made explicit) over proposing mechanisms. When tempted to add a new game-side bullet to an `Aspirational targets` section that commits the engine to a specific feel, ask whether the commitment is genuinely engine-pressure (it earns the bullet) or game-design preference (it lives in concepts). Engine-side discipline remains the standard iterative-depth shape — game-side gets the asymmetric carve-out.

**Reasoning**: Game design lives downstream of "what feels good in play," which paper alone cannot decide; premature mechanism commitment closes off optimization the player would benefit from. Engine architecture has more anchors in physical and computational constraint, so iterative depth there is less prone to over-locking. The asymmetric treatment is honest about the different evaluation surfaces — engine choices can often be reasoned about in isolation; game choices typically cannot.

**Alternatives**: (1) Apply iterative-depth uniformly across engine and game — rejected because the user explicitly differentiated; treating them alike would either over-lock game-side or under-deepen engine-side. (2) Defer all game-side concept depth indefinitely — rejected because some game design questions will eventually need depth (e.g., when a vertical slice forces a decision); the carve-out is "default malleable longer," not "never deepen." (3) Lift the same caution to engine — rejected because engine constraints are more deterministic and benefit from earlier depth.

---

## D-015 — No memory-tier for this project; all context lives in the repository — 2026-05-02

**Scope**: project

**Context**: During R-004's closeout, after a feedback memory was saved to `~/.claude/projects/-home-jaime-code-curiosity/memory/feedback_game_design_iteration.md` to capture the game-side iteration emphasis later codified as D-014, the user surfaced that memory-tier persistence is itself anti-pattern for this project: *"we should avoid using memory-tier altogether because it escapes the contextual encoding of the repository."* Memory escapes the repo's cross-machine surface — context saved to memory is not recoverable from a fresh clone of the workspace, breaking single source of truth (D-003) and cross-machine sync (D-006, D-008) the same way GitHub Issues would.

**Decision**: The auto-memory system is not used for this project. All context — user preferences, project state, behavior shaping, decisions, conventions — lives in the repository (`.claude/CLAUDE.md`, `.claude/behavior/`, `.claude/skills/curiosity/`, `history/`). When guidance worth remembering surfaces, it is routed by kind: durable reasoning into `history/decisions.md`, operational rules into `.claude/CLAUDE.md` or `.claude/behavior/`, project workflow into `.claude/skills/curiosity/SKILL.md`. Memory writes never happen in this workspace. Workspace `.claude/CLAUDE.md` carries the operational directive alongside this entry, in the same pattern as SKILL.md's no-GitHub-Issues directive aligned with D-008.

**Reasoning**: Single source of truth (D-003) and cross-machine sync via the workspace repo (D-006, D-008) require all context to be repo-embedded. Memory-tier persistence creates a parallel surface that escapes both — its contents are invisible to git, do not move with the workspace clone, and cannot be reasoned about from inspection of the repository. The same rationale that excluded GitHub Issues (D-008) excludes auto-memory.

**Alternatives**: (1) Use memory selectively for "personal preferences not specific to project context" — rejected because the boundary is fuzzy in practice (the very memory that triggered this decision was project-specific despite being framed as a preference) and clean exclusion is easier to enforce. (2) Use memory but mirror entries into the repo — rejected as duplication that violates single source of truth and creates drift opportunities. (3) Defer the policy until memory caused actual confusion — rejected because the precedent set by D-008 makes the call obvious now.

---

## D-016 — Engine task declaration surface encodes read/write, direct vs. deferred mutation, and task type — 2026-05-03

**Scope**: engine

**Context**: First-pass exploration of ECS architecture (R-005) walked the user's pre-existing service-architecture mental model (queries, commands, tasks, events) onto ECS analogs and surfaced three commitments at the surface where game and engine code declares units of scheduled work. ECS frameworks differ on what they encode at the API and what they leave informal: Bevy encodes read/write at the type level (via Rust's borrow checker) and splits direct access from a deferred command buffer; Unity DOTS encodes more (Burst-compiled jobs distinct from systems, EntityCommandBuffer for structural changes); flecs distinguishes systems from observers as separate primitives. The session needed to settle which of these the engine encodes and which it leaves to convention.

**Decision**: The engine's task declaration surface encodes three things:
1. **Read/write per component type.** Tasks declare which components they read and which they write; the runtime uses these declarations to prove which tasks can run in parallel.
2. **Direct vs. deferred mutation.** Direct access (a borrowed reference to component data) for data mutation; a deferred command-buffer-style queue for structural mutation (spawn, despawn, attach, detach).
3. **Task type.** Every task declares its type — per-frame, fixed-step, every-N, conditional. The runtime builds the schedule from task type plus read/write declarations.

Events are not encoded as a separate primitive. They remain a runtime-provided primitive (channels with writers and readers) that any task can use, but no special task type wraps them. Their concrete shape (auto-clear vs. broadcast, observer-style vs. polled, queued vs. reactive) is left to be settled by experimentation rather than committed up front.

**Reasoning**: Each of the three encoded commitments has a forcing function. Read/write at the API is unconditionally required for safe parallel scheduling — without it, the runtime cannot prove which tasks can run concurrently against shared component data. The direct-vs-deferred split is forced by the storage layer's iteration constraints regardless of which storage strategy is eventually chosen (`concepts/engine/ecs-storage.md`); structural mutations must be deferred to safe points to preserve iterator validity. Task type as metadata avoids the bigger ECS-framework anti-pattern of distinguishing many task base classes by role (Unity DOTS-style); cadence is a property of when a task runs, not what kind of object it is. Leaving events informal preserves room to choose between Bevy-style event channels and flecs-style observers (or a hybrid) once the engine has enough use cases to inform the choice.

**Alternatives**: (1) Encode less — leave the direct-vs-deferred split as a runtime convention rather than an API surface. Rejected because storage iteration forces the split anyway; encoding it makes mistakes impossible rather than merely strongly discouraged. (2) Encode more — distinguish many task base classes by role (query-task, command-task, event-task) à la Unity DOTS. Rejected as ceremony without commensurate benefit for a solo-dev engine; the Bevy approach (one task type, characteristics emerge from declared queries) is simpler and more flexible. (3) Encode events as a separate first-class primitive at the declaration surface. Rejected because events are general decoupling rather than a distinct task class; production frameworks differ on this and the choice is better left to experimentation.

---

## D-017 — Engine API uses "Task" terminology rather than ECS-framework convention — 2026-05-03

**Scope**: engine

**Context**: D-016 commits the engine to a system-of-scheduled-work declaration surface. Production ECS frameworks each have their own vocabulary for the equivalent concept — Bevy uses "System" plus "Schedule" for cadence groups, Unity DOTS uses "ComponentSystem" plus "ComponentSystemGroup," flecs uses "system" plus "phase." The session's exploration (R-005) was framed in part by mapping the user's pre-existing service-architecture mental model (queries, commands, tasks, events) onto ECS analogs; the cadence concept in particular found a natural home in the user's existing word "task." A naming choice was needed before D-016 could be expressed in the engine's vocabulary.

**Decision**: The engine's unit of scheduled work is named **Task**, with subtypes for cadence (per-frame, fixed-step, every-N, conditional). This diverges from production-framework convention. "System" is reserved for broader engine-level concerns (the audio system, the storage system, etc., as already used in `design/engine/runtime.md`) rather than for individual scheduled-work units.

**Reasoning**: This engine is explicitly a learning artifact for its author (D-001's framing of the engine as primary, reference-game as secondary; SKILL.md's "as much a learning project as a building project"). The API's primary reader is the author for the foreseeable future. Vocabulary that matches the author's mental model wins over vocabulary that matches framework convention. The cost is mild — future contributors familiar with Bevy/Unity DOTS/flecs will need a short translation table — and is acceptable given the project's posture. A secondary benefit: separating "Task" (a scheduled work unit) from "System" (a broader engine subsystem like audio or storage) reclaims "System" for the higher-level usage `runtime.md` already employs (e.g., outer-tier members described as systems). This keeps the workspace's vocabulary internally coherent.

**Alternatives**: (1) Adopt Bevy's "System" terminology. Rejected because it collides with the engine's existing "system" usage for inner/outer-tier subsystems and would force a rename of one or the other. (2) Adopt Unity DOTS's "ComponentSystem" or flecs's "system" plus "phase" terminology. Rejected for the same collision plus more ceremony in the names. (3) Coin a new term (e.g., "Worker," "Routine," "Step"). Rejected because it invents vocabulary the author has no mental model for, losing the alignment with the user's existing "task" intuition that motivated the choice.
