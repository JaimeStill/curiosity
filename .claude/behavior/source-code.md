# Source code

Boundary between source code (yours to author) and supporting
artifacts (mine to produce); checkpoint format when source code is
the next step; screen budget and phasing for larger files.

## The boundary

Source code is yours to author. Supporting artifacts are mine to
produce — and you remain hands-off on them.

**Source code (yours):**
- Function and method bodies
- Struct definitions and their fields
- Type declarations
- Variable and constant declarations within packages
- The actual implementation logic of the engine and game

**Supporting artifacts (mine, exclusively):**
- Godoc comments on exported symbols
- Inline source-code comments
- `doc.go` package documentation
- Unit tests and test infrastructure
- Scaffolding templates (e.g., `code/templates/`)
- Design documents (`design/`)
- Decisions log entries (`history/decisions.md`)
- Reset entries (`history/resets.md`)
- Conventions and configuration
- README and asset references

**Source design drives documentation.** Godoc, tests, and comments
respond to the source code as written; they never drive it. I do not
propose tests or godoc shapes that nudge the source design. These
artifacts are written or revised only during the closeout phase (per
`execution.md`), after the session's source code is in place.

## Source-code checkpoint format

When source code is the next step, my output has a defined shape:

1. **Scene-setting prelude.** What file, where in the file, what
   surrounding contracts the code participates in, what the code is
   reaching for. Enough that you can hold the change in mind before
   typing it.

2. **Source-only output.** The source code itself, in a fenced code
   block, containing only source code — no inline comments, no
   bundled tests, no godoc above. If a comment or test is needed,
   it lands as its own checkpoint.

3. **Stop.** I do not chain into the next checkpoint. After the
   source-code output, I wait.

Godoc and tests for the same symbol are produced during closeout
only, after the session's source code is in place. They do not
appear in development-phase checkpoints.

## Size and phasing

The screen budget is 30–40 lines per source-code checkpoint. This
matches the vertical space of a terminal pane at standard zoom on a
3440×1440 monitor, leaving margin for prompts and status lines.
Output that exceeds the budget is rushing past a natural pause
point.

When a file's source content exceeds the budget, break it into
phases. Phase boundaries align with natural seams in the code —
one type's methods, one related cluster of functions, one logical
unit — not arbitrary line counts. Each phase stands on its own as
a typeable unit.

Common phase patterns:

- **By type.** Phase 1: type definition and constructor. Phase 2:
  read methods. Phase 3: write methods.
- **By logical group.** Phase 1: parsing. Phase 2: validation.
  Phase 3: serialization.
- **By dependency order.** Inner types and helpers before outer
  types that compose them.

When unsure how to phase, ask before starting.
