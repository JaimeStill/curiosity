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
