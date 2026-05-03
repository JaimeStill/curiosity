# Resets Log

Append-only record of context resets — bookkeeping transactions that bring design documentation and code back into alignment. Each entry captures what triggered the reset, what design context was integrated into code or culled as obsolete, what remains forward-looking, and which decisions were promoted to the decisions log during the reset.

---

## R-001 — Workspace initialization — 2026-05-02

**Scope**: project

**Trigger**: First-time bootstrap — workspace was empty.

**Integrated**:
- `SKILL.md` (verbatim from planning conversation draft).
- `resources/assets.md` (verbatim from planning conversation draft).
- `code/conventions.md` (distilled from analysis of `~/tau/{protocol,format,provider,agent,orchestrate,examples}` and `~/code/herald`).
- `code/templates/` (six scaffolding files: `doc.go.tmpl`, `errors.go.tmpl`, `package_test.go.tmpl`, `CHANGELOG.md.tmpl`, `module.gitignore.tmpl`, `go-module-init.md`).
- Workspace `.claude/CLAUDE.md` master index plus five `.claude/behavior/` files (`execution.md`, `source-code.md`, `communication.md`, `collaboration.md`, `verification.md`).
- Nine seed entries in `history/decisions.md` (D-001 through D-009).

**Culled**: None — initialization, not a reset.

**Retained**: The `design/` subtree is intentionally not created. `design/engine/` and `design/game/` will be initialized by future sessions when their contents are written, not as empty placeholders.

**Decisions promoted**: D-001 through D-009.

**Remote**: `github.com/JaimeStill/curiosity` (source control only — no Issues, Projects, Discussions, or Wiki used).

**Follow-up**: Next session drafts `design/engine/runtime.md` on a topic branch per the workflow established in `.claude/behavior/execution.md`. The initialization session is a one-time exception to that workflow — bootstrap is committed directly to `main` since `main` does not yet exist on the remote; the topic-branch-and-PR workflow applies starting the next session.

---

## R-002 — Workflow protocol revision: design/concepts split, post-session planning, experiments — 2026-05-02

**Scope**: project

**Trigger**: First post-bootstrap session (intended to draft `design/engine/runtime.md`) surfaced three structural gaps in the workflow before drafting could proceed: (1) mixed claim-quality content inside design docs would recreate the documentation drift the workflow exists to prevent; (2) no symmetric session-end planning phase, leaving documentation decay reliant on noticing drift later rather than running every session; (3) no sanctioned location for hands-on validation of concepts that paper alone cannot settle. These were resolved through plan-mode alignment, and the resulting protocol revisions land in this session as its primary output.

**Integrated**: None — no source code was produced this session.

**Promoted**: None — the `concepts/` directory does not yet exist; no concepts were available to promote.

**Culled**:
- "Initial State" section of `SKILL.md` — removed as stale. The project is past initialization; R-001 captures what bootstrap involved, and the trajectory belongs in the reset log rather than in steady-state skill guidance.

**Retained**:
- Drafting `design/engine/runtime.md` shifts to the next session, against the new structure.
- The `concepts/` and `experiments/` directories are intentionally not materialized as empty placeholders; they appear when their first inhabitants land (mirrors R-001's approach for `design/`).

**Decisions promoted**: D-010 (design surface split: `design/` for codified intent, `concepts/` for unvalidated candidates; promotion via reset bookkeeping), D-011 (post-session planning discussion as closeout phase; reset-entry shape grows `Promoted` and `Next session focus`; closeout sequence revised), D-012 (`experiments/` directory at workspace root; exploratory R&D session type; conventions exemption; no-graduation rule).

**Next session focus**: Draft `design/engine/runtime.md` as a *context*-type session on topic branch `design-engine-runtime`. The document carries only design-grade material — claims grounded in decision-log entries (especially D-002), source code (none yet), or hard external constraints. Material that is unsettled — scheduler shape, render-thread model, ECS storage strategy, outer-tier contract specifics — is captured under `concepts/engine/` rather than embedded inside runtime.md as "Open questions." Section breakdown is settled at session start; the working shape is tier criterion, locked-in inner-tier members, locked-in outer-tier members, and runtime broad role-naming.

**Closeout-protocol observations (first run)**:
- Step ordering for context-only sessions favors decisions before docs/context updates, because the docs being adjusted (e.g., `SKILL.md`, `execution.md`) embed the decisions being recorded. The closeout sequence in `execution.md` reflects this — decisions at step 4, documentation/context at step 5. For development sessions where decisions describe new commitments and docs/context handles decay, the ordering is independent and the same sequence still applies cleanly.
- The post-session planning discussion (step 3) and the session's own opening plan-mode discussion can blur for context-type sessions, since the session's substance is itself planning. For this session, the rolling design discussion served as both — no separate step-3 pass was needed before producing closeout artifacts. Worth observing on the next context session whether the same pattern holds.

---

## R-003 — `design/engine/runtime.md` and `concepts/engine/` stubs — 2026-05-02

**Scope**: engine

**Trigger**: Carryout of R-002's "Next session focus" — first occupant of `design/engine/` plus companion stubs in `concepts/engine/` for the unsettled material runtime.md elides.

**Integrated**:
- `design/engine/runtime.md` — first design-grade engine artifact. Sections: tier criterion, inner-tier members (ECS, voxel data, physics, rendering), outer-tier members (audio, UI, storage, networking, content pipeline), runtime roles (frame loop, scheduling, lifecycle, resource ownership). Grounded in D-002. Member sections at 2–3 sentences each: responsibility plus tier-placement reasoning.
- `concepts/engine/scheduler.md`, `concepts/engine/render-thread.md`, `concepts/engine/ecs-storage.md`, `concepts/engine/outer-tier-contract.md` — each in tightest shape (question plus constraints from D-002 and runtime.md). No candidate framings; depth added when each topic becomes imminent.
- SKILL.md adjustments embedding D-013: the descriptions of `runtime.md`, `components.md`, and `components/<name>.md` revised to name the holistic-vs-index split, frame `components.md` as an index pointing into per-component depth files, and align `components/<name>.md` with the design/concepts split (unsettled material lives in `concepts/engine/`, not in component depth files).

**Promoted**: None.

**Culled**: None.

**Retained**:
- The four concept files are forward-looking; depth pass deferred until imminence.
- `design/engine/components.md` is not yet materialized; its first row appears when the first per-component depth file is created.

**Decisions promoted**: D-013 (role separation between `design/engine/runtime.md` and `design/engine/components.md`: holistic vs. index).

**Next session focus**: Draft `design/game/premise.md` as a *context*-type session on topic branch `design-game-premise`. The premise stays short (a page or two) and aspirational per SKILL.md — its job is to provide design pressure for engine validation and a coherent target for game-side work, not to specify the game. Section shape is settled at session start.

---

## R-004 — `design/game/premise.md` and `concepts/game/` first occupants — 2026-05-02

**Scope**: game

**Trigger**: Carryout of R-003's "Next session focus" — first occupant of `design/game/` plus companion stubs in `concepts/game/` for the mechanism-level material premise.md elides. Mirrors the engine-side pattern R-003 set with `runtime.md` and the four engine concept stubs.

**Integrated**:
- `design/game/premise.md` — first design-grade game artifact. Four sections: Capsule, Setting, Player, Aspirational targets (seven bullets: Fidelity, Scale, Adversity, Embodiment, Persistence, Emergence, Play). Aspirational, not specification; mechanism-level material elided to `concepts/game/`. Refinements applied during drafting: "earth-like" (not literal Earth); residual organic life beat ("contested overlay, not sterile relic"); Halo:CE vibe anchor named explicitly for emergence; "technology-now-life" framed as the load-bearing image; tension paragraph in Player; systemic-adversary sentence in Setting; "challenging without punishing" qualifier on Adversity; physics-as-gameplay-substrate captured as the **Play** bullet.
- `concepts/game/vessel-embodiment.md`, `concepts/game/building-and-research.md`, `concepts/game/character-augmentation.md`, `concepts/game/conflict-and-encounters.md`, `concepts/game/physics-playground.md` — five concept stubs in tightest shape (Question + Constraints with citations into premise and decisions log). Cross-references between stubs name the open coordination questions (vessel↔building, vessel↔augmentation, vessel↔conflict, all three↔playground) without proposing mechanisms. Mirrors the four engine concept stubs created in R-003.
- Workspace `.claude/CLAUDE.md` gained a **Memory** section per D-015 — the auto-memory system is not used for this project; all context lives in the repository.

**Promoted**: None — all five game concepts are brand new this session; none ready for design promotion.

**Culled**:
- The feedback memory file `~/.claude/projects/-home-jaime-code-curiosity/memory/feedback_game_design_iteration.md` and its companion `MEMORY.md`, plus the `memory/` subdirectory itself. Saved earlier in the session before D-015 was decided; substance preserved as D-014. Cleanup is a direct consequence of D-015.

**Retained**:
- The five game concept stubs are all forward-looking; depth pass deferred until imminence (especially per D-014, which carves out asymmetric malleability for game-side concepts).
- `design/game/` may grow additional documents as the reference game's design surfaces concrete needs; none yet.

**Decisions promoted**:
- D-014 (game-side concepts default malleable longer than the iterative-depth principle alone implies — project scope; surfaced when the user asked to optimize for what feels best rather than locking in early).
- D-015 (no memory-tier for this project; all context lives in the repository — project scope; surfaced after a feedback memory was saved earlier in the session, prompting the user to flag memory itself as anti-pattern given the single-source-of-truth and cross-machine-sync principles already established by D-003, D-006, D-008).

**Next session focus**: First per-component depth: **ECS**. Context-type session on topic branch `design-engine-components-ecs`. Drafting `design/engine/components/ecs.md` and initializing `design/engine/components.md` as the index (its first row). ECS is the inner-tier substrate per `design/engine/runtime.md` (Inner-tier members, ECS); depth here informs every other inner-tier component. Concept material from `concepts/engine/ecs-storage.md` likely promotes during this work — promotion is bookkeeping in the closeout reset, not assumed up front. Section shape settled at session start.

**Closeout-protocol observations (third run)**:
- The post-session planning discussion surfaced two architectural decisions (D-014, D-015) that emerged from later turns rather than from drafting the documents themselves. Both came out of user-facing meta-feedback ("don't lock in"; "no memory-tier") rather than from the planning surface itself. The closeout sequence handled this cleanly — decisions logged before doc updates (R-002's observation held), with the consequences of D-015 (memory cleanup, CLAUDE.md addition) cascading into step 5.
- Mid-session course corrections extended the agenda twice (conflict-and-encounters, physics-playground). Execution.md's wording — "the agenda is a forward commitment that adjusts as we work; it is not a contract" — held up. Both extensions stayed within session type (context) and topic (design-game-premise), so no branch change or session reset was needed.
