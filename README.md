# curiosity

Workspace for a Go-based voxel game engine project (codename `curiosity`).

## What this is

The durable, cross-machine context for the project: the project skill, design
docs, decision history, conventions, templates, and asset references. This
repository persists the planning surface and behavior layer across machines.

## What this is not

Not engine source code. Not game source code. Sub-projects (engine modules,
the reference game) live in their own git repositories under subdirectories
that are `.gitignore`d from this repo (`/engine/`, `/game/`, etc.).

## Onboarding

Open this directory in Claude Code. The curiosity skill at
`.claude/skills/curiosity/SKILL.md` auto-loads and is the entry point for the
project workflow. The workspace `.claude/CLAUDE.md` encodes Claude Code
behavior in this workspace.

For project context, start with `.claude/skills/curiosity/history/resets.md`
(latest entry shows where the prior session left things) and
`.claude/skills/curiosity/history/decisions.md` (architectural decisions log).
Design docs (when they exist) live under `.claude/skills/curiosity/design/`.

## Source control

This workspace is tracked at `github.com/JaimeStill/curiosity`. GitHub is used
**for source control only** — no Issues, Projects, Discussions, or Wiki for
planning. All planning context lives in the repository so the single source
of truth is preserved.

## Project summary

The primary artifact is a generic Go-based voxel game engine, designed to
support modern voxel-based games at high fidelity. A reference game is
developed alongside the engine to keep one concrete scenario sharp enough to
surface real engine requirements; engine decisions serve the generic case,
not premise-specific needs. See `.claude/skills/curiosity/SKILL.md` for full
project workflow.
