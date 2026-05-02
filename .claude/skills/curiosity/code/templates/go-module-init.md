# Module Initialization Runbook

How to spin up a new Go module in the curiosity engine workspace.

## Steps

1. Create the module directory under the engine repository:

       mkdir -p <module-name>
       cd <module-name>

2. Initialize the Go module:

       go mod init <module-path>
       go mod edit -go=1.26.0

3. Place starter files from `code/templates/`, replacing placeholders:

   - `errors.go` — copy from `errors.go.tmpl`. Substitute `{{PACKAGE}}` and `{{Type}}`.
   - `tests/<package>/<package>_test.go` — copy from `package_test.go.tmpl`. Substitute `{{PACKAGE}}`, `{{MODULE_PATH}}`, and `{{Type}}`.
   - `CHANGELOG.md` — copy from `CHANGELOG.md.tmpl`. Substitute `{{MODULE_NAME}}` in the example release block when ready.
   - `.gitignore` — copy from `module.gitignore.tmpl` verbatim.

4. **Defer `doc.go` until the package is considered complete** (per `code/conventions.md` §4). Writing `doc.go` over volatile infrastructure creates documentation drift. The template at `doc.go.tmpl` is invoked during the closeout phase of the session in which the package reaches stability — not as part of new-package scaffolding.

5. If this module lives directly under `~/code/curiosity/` (e.g., as a sub-project repository), add its directory name to the workspace `.gitignore` so the workspace repo does not try to track it.

6. The new module repository gets its own `git init` independent of the workspace. Each sub-project's git history is isolated.

## References

- `code/conventions.md` — overall Go style for this project.
- `~/tau/agent/` and `~/tau/orchestrate/` — reference module shapes.
- `~/code/herald/` — reference for a single-module Go application.
