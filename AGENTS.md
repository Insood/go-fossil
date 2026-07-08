# Agent Notes

This repo is for a small 3D salvage game built in Go and raylib.

For the higher-level layout and implementation guidance, see [docs/technical-direction.md](docs/technical-direction.md) and [docs/llm-context.md](docs/llm-context.md).

## Development Rules

- Keep the codebase small and understandable.
- Prefer straightforward implementations over clever abstractions.
- Build gameplay in small, playable slices.
- Keep rendering, terrain, and gameplay systems separated where practical.
- Prefer clear, testable code.
- Avoid adding dependencies unless they solve a real problem.
- Preserve existing user changes unless explicitly asked to modify them.
- If a change affects gameplay structure, explain the tradeoff briefly.
- When you generate code, update docs if the change affects terrain behavior, camera behavior, system ownership, or gameplay rules.

## Change Placement

- Put gameplay code in `cmd/game` by default.
- Put ECS components in `cmd/game/components.go`.
- Put one system per file in `cmd/game/*_system.go`.
- Put rendering changes in the render system unless the feature clearly needs a separate pass.
- Put constants and tuning values in `cmd/game/config.go`.
- Put shared math and utility helpers in `cmd/game/utils.go`.
- Put asset loading and cleanup in `cmd/game/game.go`.
- Keep `main.go` as small as possible.
- Put terrain-specific domain rules in `internal/terrain` if they outgrow the executable package.

## Behavior Invariants

- The scene is fundamentally 3D.
- The camera is orthographic and should support an isometric-style presentation.
- Terrain begins as a heightmap-derived mesh.
- The salvage loop depends on readable terrain shape and embedded-object visibility.
- Early implementations may fake or simplify cutting behavior if that helps validate the loop faster.

## Do Not

- Do not move core gameplay logic into `main.go`.
- Do not add new dependencies unless they solve a real problem.
- Do not split systems into extra packages unless the repo clearly needs it.
- Do not harden terrain architecture too early.
- Do not change major world assumptions without updating the docs.
