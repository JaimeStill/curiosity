# Engine components

Index of per-component depth specifications. Each entry points to a
depth file in `components/`. Components without depth files are not
listed here; their tier placement and high-level responsibility are
named in `runtime.md`. Per D-013, this index carries the inventory
of depth files only, not tier framing or interface shape (interface
shape is concept-tier until firm per D-010).

New entries land when a component receives a depth file — typically
because it is about to receive concrete attention; see SKILL.md.

## Inner tier

- **ECS** — [`components/ecs.md`](components/ecs.md). Entity,
  component, and storage primitives, plus the typed call-site
  surface the engine consumes.
