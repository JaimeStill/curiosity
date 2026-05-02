# Execution

How work flows through a session in this workspace. Loaded when
starting work, structuring a checkpoint, deciding tool spread, or
preparing to commit or push.

## Session lifecycle

### Session start

In plan mode at the start of a session, two things happen before we
exit it:

1. **Orient.** Read the most recent entry in `history/resets.md` to
   see where the prior session left things — what was integrated,
   what remains forward-looking. On a fresh project where `resets.md`
   holds only the initialization entry, that entry IS the
   orientation. Load relevant skill content.

2. **Scope.** Lay out the session agenda — what we're aiming to land,
   scoped to a token budget that leaves headroom against context rot
   while still completing meaningful work. The agenda is a forward
   commitment that adjusts as we work; it is not a contract.

Plan mode exits only after both have happened.

### Branch creation

Immediately after plan mode exits, create a branch. Branch name is a
topic slug derived from the session agenda — e.g., `init-workspace`,
`voxel-mesh-naive`. Short, lowercase, hyphen-separated. All session
work happens on the branch.

### Development phase

Source-code design and implementation. You author per
`source-code.md`; I provide scene-setting, scaffolding, and
surrounding context. The development phase ends when the source code
for the session agenda is implemented and we agree to transition.

Pacing and checkpoint rules (below) apply throughout.

### Closeout phase

My responsibilities engage in coordinated order:

1. **Godoc comments.** Write or revise godoc on new and modified
   exported symbols.
2. **Unit tests.** Write or revise tests covering the session's
   source-code changes.
3. **Documentation and context artifacts.** Update `.claude/`
   infrastructure as needed: design docs that have been absorbed
   into code get culled (per the documentation-decay discipline in
   `SKILL.md`); decisions made during the session get appended to
   `history/decisions.md`; a reset entry gets appended to
   `history/resets.md`.
4. **Commit.** Ask before committing. Commit message follows the
   style in `~/.claude/CLAUDE.md`: imperative subject, body paragraph.
5. **Push.** Ask before pushing.
6. **PR.** Create a pull request using `gh pr create`. The PR body
   is lifted from the closeout entry just appended to `resets.md` —
   no separate authoring, no drift between the durable record and
   the PR description.

The session is complete once the PR is submitted. Post-PR cleanup
(merge, branch deletion) is yours.

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
