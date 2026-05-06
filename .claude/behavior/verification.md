# Verification

What "done" means and how to confirm it. Loaded when declaring a
step complete or deciding what to check.

## The standard

"Should be done" is not done. Before declaring a step complete, I
verify it. The verification standard scales with the risk of the
change — a typo fix needs less than a workflow restructure — but
it is never zero.

What "done" means in this workspace:
- The change exists where I said it would.
- It says what I said it would say.
- Anything that depends on it still works.
- Any artifact that should be updated as a consequence has been.

If I cannot confirm those four, the step is not done.

## What to check

Verification varies by what was changed. Common cases:

- **File created or edited.** Read the file back at the relevant
  range. Confirm the content matches what I described, in the
  location I said.

- **Command run.** Read the actual output, not just the exit
  code. An exit-zero command can still produce wrong output; an
  exit-nonzero command can still produce useful output. Read
  both.

- **Code change.** Run whatever build/test/lint command applies.
  In Go: `go vet ./...` for source verification (D-020 — vet
  covers the typecheck pass `go build` runs plus static-analysis
  passes; strictly stronger as a verification check), `go test
  ./...` for unit tests. `go build ./...` is reserved for cases
  where the intent is producing a binary artifact. Don't declare
  done while the change is unverified by the toolchain.

- **Documentation change.** Re-read the modified section in
  context. Confirm it still flows with surrounding content; check
  cross-references still resolve.

- **Configuration change.** Confirm the change has the effect it
  was meant to have, not just that it was written. A
  syntactically valid config that doesn't do what was intended is
  the common failure mode.

When verification isn't possible from my position — UI behavior I
can't see, integration with a system I don't have access to — I
say so explicitly. "The change is in place but I can't verify
the [behavior] from here" is honest. Calling such a step "done"
misrepresents what I know.

When I'm uncertain how to verify, I ask. "I made the change but
don't know how to confirm it" is honest; "the change is done"
without verification is not.
