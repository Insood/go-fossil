# Agent Notes

This repo is for a small 3D salvage game built in Go and raylib.

For the higher-level layout and implementation guidance, see [docs/technical-direction.md](docs/technical-direction.md) and [docs/llm-context.md](docs/llm-context.md).

## Development Rules

- Keep the codebase small and understandable.
- Prefer straightforward implementations over clever abstractions.
- Build gameplay in small, playable slices.
- Keep rendering, terrain, input, UI, and gameplay systems separated where practical.
- Prefer clear, testable code.
- Favor battle-tested libraries over custom NIH code when they materially simplify implementation.
- Prefer dependencies already present in `go.mod` before adding anything new.
- Ask for permission before adding a new dependency.
- Preserve existing user changes unless explicitly asked to modify them.
- If a change affects gameplay structure, explain the tradeoff briefly.
- When you change behavior, update the existing docs that describe that behavior in the same change.
- When a behavior change does not fit the current docs, add focused documentation for it and link it from the relevant LLM-facing doc.

## Change Placement

- Put gameplay code in `cmd/game` by default.
- Put ECS components in `cmd/game/components.go`.
- Put one system per file in `cmd/game/*_system.go`.
- Put rendering changes in the render system unless the feature clearly needs a separate pass.
- Put constants and tuning values in `cmd/game/config.go`.
- Put shared math and utility helpers in `cmd/game/utils.go`.
- Put asset loading and cleanup in `cmd/game/game.go` and `cmd/game/asset_manager.go`.
- Keep `main.go` as small as possible.
- Keep terrain chunk parsing, validation, mesh data generation, baked terrain texture generation, and terrain sampling helpers in `internal/terrain`.

## Behavior Invariants

- The scene is fundamentally 3D.
- The main camera is orthographic and supports an isometric-style presentation.
- Terrain chunks are heightmap-derived meshes.
- Cutting is currently represented with artifact ID masks and overlay textures; height samples are not mutated by the laser.
- The salvage loop depends on readable terrain shape, embedded-object visibility, drone viewport aiming, fragment scoring, and generated chunk expansion.

## Do Not

- Do not move core gameplay logic into `main.go`.
- Do not add new dependencies without explicit user permission.
- Do not split systems into extra packages unless the repo clearly needs it.
- Do not introduce speculative systems for inventory, progression, enemies, hazards, or AI.
- Do not change terrain behavior, camera behavior, system ownership, system order, gameplay rules, or major world assumptions without updating the docs.
