# Execution

How work flows through a session in this workspace. Loaded when
starting work, structuring a checkpoint, deciding tool spread, or
preparing to commit or push.

## Session types

Three session types are recognized; type is named in the session
agenda at session start and confirmed (or revised) by the
post-session planning discussion that closes the prior session.

- **Development sessions** produce engine or game source code. They
  have a development phase between session start and closeout. Godoc
  and unit tests are written during closeout.
- **Context sessions** produce or revise planning-surface artifacts
  (design, concepts, history, behavior layer, conventions, skill
  content). They have no development phase — work happens directly
  between session start and closeout. Godoc and unit-test steps in
  closeout are no-ops.
- **Experiment sessions** produce throwaway code under
  `experiments/<name>/` together with a `README.md` recording the
  finding. The development phase produces the experiment. Godoc and
  unit-test steps in closeout are no-ops — experiments are exempt
  from `code/conventions.md`.

## Session lifecycle

### Session start

In plan mode at the start of a session, two things happen before we
exit it:

1. **Orient.** Read the most recent file in `history/resets/` (sorted
   lexically by filename — the highest `R-###.<slug>.md`) to see where
   the prior session left things — what was integrated, what was
   promoted, what was culled, what remains forward-looking, and what
   the prior session named as the next session's focus and type. Load
   relevant skill content.

2. **Scope.** Lay out the session agenda — type, target, and a token
   budget that leaves headroom against context rot while still
   completing meaningful work. The agenda is a forward commitment
   that adjusts as we work; it is not a contract.

Plan mode exits only after both have happened.

### Branch creation

Immediately after plan mode exits, create a branch. Branch name is a
topic slug derived from the session agenda — e.g., `init-workspace`,
`voxel-mesh-naive`, `workflow-protocol-revision`. Short, lowercase,
hyphen-separated. All session work happens on the branch.

### Development phase

Source-code design and implementation (development sessions) or
prototype construction (experiment sessions). You author per
`source-code.md`; I provide scene-setting, scaffolding, and
surrounding context. The phase ends when the session's intended
source is in place and we agree to transition.

Context sessions skip this phase entirely — the planning-surface
work happens directly between session start and closeout.

Pacing and checkpoint rules (below) apply throughout.

### Closeout phase

My responsibilities engage in coordinated order:

1. **Godoc comments.** Write or revise godoc on new and modified
   exported symbols. *No-op for context and experiment sessions.*

2. **Unit tests.** Write or revise tests covering the session's
   source-code changes. *No-op for context and experiment sessions.*

3. **Post-session planning discussion.** Conducted in plan mode.
   Review what the session accomplished. Evaluate the current state
   of `design/`, `concepts/`, and `experiments/`. Surface (a) any
   concept→design promotions, (b) any concept culls, (c) any design
   absorptions where source landed, (d) any experiment removals
   whose job has ended, (e) the next session's focus and type. The
   discussion produces the to-do list for the remaining closeout
   steps.

4. **Decisions log.** Add new entries to
   `.claude/skills/curiosity/history/decisions/` (one file per decision,
   `D-###.<slug>.md`) for any architectural decisions surfaced during
   the session, in the shape defined in `SKILL.md`'s Decisions Log
   section.

5. **Documentation and context artifacts.** Apply the outcomes from
   step 3 against the planning surface: update `design/` and
   `concepts/`; move promoted concept files into the design tree;
   remove culled concepts and finished experiments; update `.claude/`
   infrastructure if it changed during the session. When a doc
   embeds a decision recorded in step 4, this is the step that lands
   that change.

6. **Reset entry.** Add a new file to
   `.claude/skills/curiosity/history/resets/` (`R-###.<slug>.md`) using
   the shape defined in `SKILL.md`'s Reset Protocol section, including
   the `Next session focus` field set from step 3.

7. **Commit / push / PR.** Ask before committing. Commit message
   follows the style in `~/.claude/CLAUDE.md`: imperative subject,
   body paragraph. Ask before pushing. Create a pull request using
   `gh pr create`; the PR body is lifted from the reset entry — no
   separate authoring, no drift between the durable record and the
   PR description.

   PR creation closes the work the branch is for, not the session.
   When the branch is single-session, the two coincide and the PR
   lands at session closeout. When the branch carries multiple
   sessions toward a single target (experiment branches,
   multi-phase refactors), commit and push happen per session and
   the PR is created when the work the branch is for completes.

The session is complete once its commit + push lands. PR
submission closes the branch's work — at the same closeout for
single-session branches, at the closing session for multi-session
branches. Post-PR cleanup (merge, branch deletion) is yours.

## Pacing and checkpoints

One checkpoint at a time. A checkpoint is a single scoped concern —
typically one file, occasionally a small group of tightly related
ones. When I'm uncertain whether something is one checkpoint or
multiple, it's multiple.

After a checkpoint lands, stop. Let it be reviewed before proposing
the next. The instinct to chain "while I'm at it" extensions is the
instinct to break — even small extensions accumulate into context
drift.

Pacing applies in any phase. Source-code checkpoints carry an
additional constraint (screen budget) detailed in `source-code.md`.

## Tool use

Direct and sequential by default. Parallel tool calls and subagents
are appropriate for genuinely independent work or broad exploration
where context-window protection matters; they are not appropriate as
a way to feel productive during methodical execution.

The test: is the work genuinely independent, or am I splitting it
for the sake of splitting? If the latter, stay sequential.

## Ending a turn

State the result. Point to the next concrete step. Stop.

No recap of what just happened when the change is visible. No
anxious restatement of what was decided. End-of-turn output is the
smallest text that orients you to "what's done, what's next."
