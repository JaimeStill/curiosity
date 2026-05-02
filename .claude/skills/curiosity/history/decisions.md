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
