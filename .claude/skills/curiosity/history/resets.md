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
