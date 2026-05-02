# Go Conventions

Go style adopted for this project, distilled from analysis of `~/tau/{protocol,format,provider,agent,orchestrate,examples}` and `~/code/herald`. Reference material — consulted at module-creation time and when surfacing an unsettled question. Not subject to the documentation-decay discipline that governs `design/`.

## 1. Scope and audience

These conventions apply to all Go code in this project — engine, game, and any tooling — with one critical exception detailed in section 2. Read section 2 before applying anything else here.

The reference codebases (tau and herald) are mature, production-grade Go and exemplify a coherent style that has held up under load. The conventions below are not invented; they are observed patterns codified for reuse.

## 2. The two-tier divergence (read this first)

This project has two tiers of code with different performance characteristics, and they follow these conventions differently.

**Outer tier** — audio, UI, networking, storage, content pipeline, save/load, asset loaders, debug tools, editor tooling, and any cold-path code. Follows these conventions verbatim.

**Inner tier** — ECS, voxel data, physics, rendering, and anything sharing the per-frame memory and lifecycle budget. Deliberately diverges from these conventions where the divergence is justified by hot-path constraints:

- No JSON-everywhere configuration. Hot-path subsystems take concrete typed parameters at construction.
- No registry indirection on hot paths. Compile-time wiring or direct calls.
- No allocation-heavy interface dispatch in tight loops. Concrete types and value receivers.
- No goroutine-per-message. Structured worker pools with fixed-size goroutine sets.

Inner-tier deviations require an inline comment pointing back to this section and naming the specific constraint. Cold-path code that happens to live in the inner tier (loaders, serializers, debug tooling) reverts to outer-tier conventions.

When in doubt, default to outer-tier. The inner-tier exception is for code where you can demonstrate the cost of conformance, not code that might one day be hot.

## 3. Multi-module workspace topology

Each major component is its own Go module with its own `go.mod`, `CHANGELOG.md`, and (when stable) `doc.go` per package. Sub-modules within a component repository isolate heavy third-party dependencies — for example, a cloud-provider integration lives in its own sub-module so consumers of the parent package don't pay for the cloud SDK transitively.

`go.work` lives at the engine repository root and unifies engine modules during local development. The workspace at `~/code/curiosity/` is **not** a Go workspace; it is the planning superproject. Engine modules live under `~/code/curiosity/engine/` (gitignored from the workspace repo, with their own git history).

Target Go version: 1.26.x. Pin in each module's `go.mod` via `go 1.26.0`.

References:
- `~/tau/format/converse/go.mod` — sub-module isolating AWS Converse SDK
- `~/tau/provider/azure/go.mod` — sub-module isolating Azure SDK
- `~/code/herald/go.mod` — single-module application targeting Go 1.26

## 4. Package documentation

Each package has a `doc.go` file containing the package-level godoc. Style is narrative-first: explain what the package is, what it provides, and how a consumer uses it. Length follows the package's surface area — small utility packages have a paragraph; orchestration packages can run 100+ lines with example blocks.

References:
- `~/tau/orchestrate/doc.go` — short overview pointing to sub-packages
- `~/tau/orchestrate/workflows/doc.go` — extensive narrative with code examples
- `~/tau/agent/doc.go` — sectioned godoc covering creation, options, errors, thread safety

**Timing.** `doc.go` is **not** written until the package or module is considered complete (API surface settled, no major refactors anticipated). Writing `doc.go` over volatile infrastructure creates documentation drift — the file has to be revised every time the package churns. Defer authoring until stability is reached; the template in `code/templates/doc.go.tmpl` exists for that moment, not before.

The `doc.go` template is invoked during the closeout phase of the session in which the package reaches stability, not as part of new-package scaffolding.

## 5. Interfaces

Defined at the consumer site, not the producer site. Implementations are unexported structs; constructors return the interface type.

- **Size.** 1–7 methods. If an interface grows beyond that, ask whether it should be split or whether the consumer actually needs all of it.
- **Naming.** Descriptive names, no `-er` suffix. `Provider`, `Format`, `StreamReader`, `Participant`, `Observer`, `StateNode` — not `Providerer` or `StreamReaderer`.
- **Methods.** Verb-noun (`Endpoint()`, `SetHeaders()`, `PrepareRequest()`). Getters omit `Get` (`Name()`, not `GetName()`).
- **Placement.** In the same file as the package's primary type, near the top.

**Inner-tier divergence.** Hot-path code prefers concrete types and value receivers over interface dispatch. Reach for interfaces only when polymorphism actually pays in measurable runtime terms. A struct method call is cheaper than an interface call; the inner tier respects that.

References:
- `~/tau/format/format.go:10-22` — `Format` interface (3 methods)
- `~/tau/provider/provider.go:14-39` — `Provider` interface (6 methods)
- `~/tau/protocol/streaming/streaming.go:24-29` — `StreamReader` (1 method)
- `~/code/herald/pkg/database/database.go:19-25` — `System` interface (2 methods)

## 6. Constructors and dependency injection

Constructors are named `New{Type}` and take config plus dependencies as explicit parameters. No service locator, no IoC container, no functional-options pattern for production code.

```
func New(cfg *Config, dep1 Dep1, dep2 Dep2) (Type, error)
```

Returns the interface type; the concrete struct stays unexported. Validation happens inside the constructor with early returns on missing or invalid inputs.

Defaults are exposed via a separate function:

```
func DefaultConfig() *Config { ... }
```

Config is **ephemeral** — it is consumed at construction to initialize the type's fields, then the config struct is not retained. This means runtime types do not carry the JSON shape of their config; they carry the typed fields they actually need.

Functional options are used in test mocks (`mock.WithChatResponse(...)`) but not in production constructors. The reasoning: production code benefits from explicit, named parameters that make the dependency graph visible at the call site. Options patterns hide it.

References:
- `~/tau/agent/agent.go:81` — `agent.New(cfg, provider, format)` shape
- `~/tau/provider/azure/azure.go:27-56` — early-return validation in constructor
- `~/tau/provider/base.go:6-29` — `BaseProvider` for embedding into concrete providers

## 7. Configuration lifecycle

Configuration follows a three-phase lifecycle:

1. `loadDefaults()` — populate hardcoded sensible defaults.
2. `loadEnv()` and/or `Merge(overlay)` — apply environment variable overrides; merge any layered configuration from disk.
3. `validate()` — confirm the resulting config is internally consistent before it is consumed.

Each config struct has a `Finalize()` method that runs the three phases in order. Sub-config structs have their own `Finalize()`, called cascadingly from the parent. Environment-variable names live in centralized `Env` structs (one per package that owns env-driven config).

Configuration is JSON-driven for outer-tier code, with string duration fields parsed via helper methods (`TimeoutDuration()` parses a `Timeout string` into `time.Duration`).

**Inner-tier divergence.** Hot-path subsystems take concrete typed parameters at construction. ECS, voxel grids, physics colliders — these don't load from JSON. They are configured by the outer tier that wraps them, which may itself load from JSON, but the inner-tier struct never sees the JSON shape.

References:
- `~/code/herald/internal/config/config.go:89-165` — three-phase `Load()` and `finalize()`
- `~/code/herald/pkg/auth/config.go:69-226` — `Finalize(env)` and `Merge(overlay)` on a sub-config
- `~/tau/protocol/config/agent.go:35-71` — layered merge across protocol/client/model

## 8. Errors

Three-layered error pattern:

**Sentinel errors.** Defined in a dedicated `errors.go` file. Used for stable error identity that callers can compare with `errors.Is`.

```
var (
    ErrNotFound   = errors.New("agent not found")
    ErrExists     = errors.New("agent already exists")
    ErrEmptyName  = errors.New("agent name is empty")
)
```

**Custom error types.** Struct-based with functional options for rich context (origin, code, cause, timestamp). Implements `Error()` and `Unwrap()`.

```
type AgentError struct {
    Type      ErrorType
    ID        string
    Code      string
    Message   string
    Cause     error
    Timestamp time.Time
}
```

Constructors take options:

```
err := NewAgentError(ErrorTypeInit, "failed to initialize",
    WithCode("AUTH_FAILED"), WithCause(innerErr), WithID(uuid))
```

**Wrapping.** Universal `fmt.Errorf("context: %w", err)` for adding call-site context to a propagating error. Every error path wraps with descriptive context.

**Zero panics.** Production code does not panic. Errors are returned, never thrown. Panics are reserved for genuinely unrecoverable invariant violations during initialization, not runtime conditions.

References:
- `~/tau/agent/errors.go` — sentinel errors and `AgentError` struct with options
- `~/code/herald/pkg/auth/errors.go:1-13` — sentinel error pattern
- `~/code/herald/pkg/repository/errors.go:1-30` — `MapError` helper for DB → domain translation
- `~/code/herald/internal/documents/errors.go:18-33` — `MapHTTPStatus` for HTTP boundary

## 9. Testing

Tests live in a sibling `tests/` directory, not co-located `_test.go` files. Test packages use the `_test` suffix: `package agent_test`, `package hub_test`. This is black-box testing — only the public API is exercised, no access to unexported helpers.

Style is heavily table-driven:

```
cases := []struct {
    name string
    in   InputType
    want WantType
}{
    {"basic case", InputType{...}, WantType{...}},
    {"edge case", InputType{...}, WantType{...}},
}
for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) {
        got := SystemUnderTest(tc.in)
        if got != tc.want { t.Errorf(...) }
    })
}
```

Helper functions use `t.Helper()` and stay in the same test file. No separate fixtures package unless reuse demands it.

Real dependencies are preferred where practical: `t.TempDir()` for filesystem, `httptest.NewServer` for HTTP, real database connections for repository tests. Mocks are used when external dependencies make real impractical, and mocks themselves use functional options for ergonomic setup (`mock.WithChatResponse(...)`).

**Inner-tier divergence.** Hot-path code may co-locate `_test.go` for benchmark access to unexported helpers. Document case-by-case why colocation is needed; the default remains sibling `tests/`.

References:
- `~/tau/orchestrate/tests/` — black-box test layout for the orchestrate module
- `~/tau/agent/mock/agent.go` — mock with functional options
- `~/code/herald/tests/config/config_test.go:109-217` — table-driven config test
- `~/code/herald/tests/handlers/handlers_test.go:15-85` — black-box handler tests

## 10. Concurrency

`context.Context` is propagated through every potentially-blocking function as the first parameter. No package-level globals for cancellation; no goroutines that lack a way to be stopped.

`sync.RWMutex` guards shared maps and registries. Reader lock for reads, writer lock for writes. No fancy lock-free schemes unless profiling demands them.

Goroutine-per-message is the outer-tier pattern: each inbound message spawns its own goroutine, no artificial pooling. The Go scheduler handles efficiency.

Lifecycle and shutdown are coordinated via cancel-context plus `sync.WaitGroup`. Components register startup and shutdown hooks; the coordinator cancels context and waits for hooks to drain.

Modern idioms used throughout:
- `sync.WaitGroup.Go(fn)` (Go 1.22+) instead of manual `wg.Add(1)` + `go fn()` + `wg.Done()`.
- `slog` for structured logging, with module-scoped loggers passed through constructors.
- `maps.Clone()` and `maps.Copy()` for map handling.
- UUIDv7 for time-sortable identifiers.

**Inner-tier divergence.** Frame-loop code uses fixed-size goroutine pools, not goroutine-per-task. Channel-free shared state with explicit memory ordering (atomics) is permitted when justified by measured contention. Cancel-context still propagates, but the per-frame budget is too tight for arbitrary goroutine spawning.

References:
- `~/tau/orchestrate/hub/hub.go:76,360-395` — message loop goroutine and per-message dispatch
- `~/tau/provider/streaming/sse.go:24-75` — context-aware streaming reader
- `~/code/herald/pkg/lifecycle/lifecycle.go` — startup/shutdown coordination

## 11. Registry pattern

When components register themselves at process start (via `init()` or explicit `Register()`), the registry uses a `sync.RWMutex`-backed map keyed by name. Read access during normal operation; write access only during registration.

The pattern:

```
type Factory func(cfg *Config) (Provider, error)

var registry = struct {
    mu        sync.RWMutex
    factories map[string]Factory
}{factories: make(map[string]Factory)}

func Register(name string, factory Factory) { ... }
func Create(name string, cfg *Config) (Provider, error) { ... }
```

Registries are **outer tier only**. Hot-path code does not pay for registry indirection.

References:
- `~/tau/provider/registry.go:14-32` — provider factory registry
- `~/tau/format/registry.go:12-25` — format factory registry

## 12. Generics

Used strategically, not as the default reach. The bar: does the type parameter remove genuine duplication that interfaces or `any` cannot handle without losing meaningful type safety?

Cases where generics earn their place in tau:
- `MessageChannel[T any]` — type-safe async channels.
- `ProcessChain[TItem, TContext any]` — generic fold over a sequence.
- `ProcessParallel[TItem, TResult any]` — generic worker pool.
- `PageResult[T any]` — generic paginated response wrapper.

Cases where generics are avoided in favor of interfaces:
- Plugin registries (use `any` for the value type; plugins are small and the type assertion happens once at lookup).
- Polymorphic content types (use a tagged interface, not a sum-type emulation).

When unsure, write the non-generic version first. Generic-ify only after a second use case shows up that justifies the abstraction.

## 13. Observability and immutability

Outer-tier state types carry an `Observer` that receives events on every mutation. The `Observer` interface is one method (`OnEvent(ctx, event)`) and ships with a `NoOpObserver` zero-cost default — observability is opt-in but pervasive.

State transitions return new state values rather than mutating in place:

```
func (s State) Set(key string, value any) State { ... }
```

Observers are notified inside the transition. Immutability prevents an entire class of subtle bugs in concurrent code; observability gives a coherent timeline of what happened.

**Inner-tier divergence.** Per-frame state cannot afford allocation per mutation. Inner-tier code mutates in place and emits observer events at frame boundaries or coarser, not per-mutation. The observer interface is the same; the granularity differs.

References:
- `~/tau/orchestrate/state/state.go:27-58` — immutable State with Observer
- `~/tau/orchestrate/observability/observer.go:70` — Observer interface

## 14. File granularity

Files stay small (~40–100 lines is typical) and are separated by responsibility within a package:

- `<name>.go` — primary type and its methods.
- `types.go` — wire-format structs (JSON-tagged) used by the package.
- `marshal.go` — encoding logic.
- `parse.go` — decoding logic.
- `errors.go` — sentinel errors and custom error types for the package.
- `config.go` — package-specific configuration struct.

Methods on a primary type can spread across multiple files when grouped by concern (e.g., `agent_chat.go`, `agent_stream.go`, `agent_tools.go`) but only when the file would otherwise exceed roughly 200 lines or mix unrelated responsibilities.

## 15. Tooling

No Makefile mandate. `mise.toml` is permitted (herald uses it for dev tasks like build, test, vet, run); per-module decisions on whether to adopt it.

No enforced `.golangci.yml` at this stage. `gofmt` and `go vet` are the baseline. Linter configuration is revisited when the engine repository is established and we have measurable signal about what lint rules add value.

`go.work` lives at the engine root only. Per-module `go.mod` pins Go version. CHANGELOG.md per module follows semantic versioning and the per-release format documented in `code/templates/CHANGELOG.md.tmpl`.

## 16. What this does NOT cover

The following are decided per sub-project, not at the workspace conventions level:

- **Release workflow** — when, how, and what triggers a versioned release.
- **CI configuration** — what runs in GitHub Actions and how it gates merging.
- **CHANGELOG conventions beyond the template** — wording style, granularity of entries, how breaking changes are flagged.
- **Module-specific naming** — public type names, error code formats, package-internal organizational patterns.

These belong in each sub-project's own conventions or documentation, evolved as the sub-project matures.
