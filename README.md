# go-fossil

`go-fossil` is a 3D salvage game built with Go and raylib.

The player controls a drone exploring a fossilized junkyard from an orthographic, isometric-style camera view. The drone searches for valuable artifacts embedded in the terrain, cuts them free with a laser, and avoids hazards such as security drones and junkyard rats.

## Core Fantasy

Fly low through a dead-tech graveyard, scan the landscape, carve out buried relics, and escape with the best finds.

## Project Direction

- Engine language: Go
- Rendering library: raylib
- World presentation: 3D scene rendered with an orthographic camera
- Terrain: mesh generated from level metadata plus a grayscale heightmap image
- Gameplay interaction: the drone can cut terrain to expose and extract embedded objects
- Entity model: ECS using [`mlange-42/ark`](https://github.com/mlange-42/ark)
- Development style: extremely small, testable implementation slices

## Near-Term Priorities

1. Establish the application shell and game loop.
2. Render a simple orthographic 3D scene.
3. Load or generate a small terrain mesh from level metadata and a heightmap image.
4. Spawn a controllable drone entity.
5. Define the first ECS components and systems.
6. Add one interactive salvageable object.
7. Prototype terrain cutting in the smallest possible vertical slice.

## Documentation Map

- [docs/project-brief.md](docs/project-brief.md): game premise, pillars, and scope guardrails
- [docs/technical-direction.md](docs/technical-direction.md): architecture choices and technical constraints
- [docs/agent-guide.md](docs/agent-guide.md): working agreements for coding agents
- [docs/llm-context.md](docs/llm-context.md): implementation map for future coding agents
- [docs/initial-slices.md](docs/initial-slices.md): suggested early implementation sequence

## Repository Shape

This project is expected to follow the same general structure as the earlier `go-towerdefense` project:

- `cmd/game` holds the executable game and most ECS-facing code
- `cmd/game/components.go` defines ECS components
- `cmd/game/*_system.go` keeps one gameplay or render system per file
- `cmd/game/config.go` stores gameplay constants and tuning values
- `cmd/game/game.go` owns bootstrapping, system wiring, and asset lifecycle
- `internal/*` packages are used sparingly for focused domain logic that deserves isolation

Current terrain content format:

- level metadata lives in `cmd/game/assets/levels/*.json`
- tile layout stays in JSON
- vertex heights come from a referenced grayscale PNG sized `(width+1) x (height+1)`
- `AssetManager` loads embedded level assets and `internal/terrain` validates/parses them
- terrain rendering now uses one mesh generated from the height samples plus one baked world texture assembled from the level tile textures

## Build

On WSL/Ubuntu, build the game with:

```bash
make build
```

The build writes the Linux binary to `bin/go-fossil` and uses a repo-local Go build cache.
Shader assets are embedded into the binary with `go:embed`, so no extra asset copy step is required for builds.

To cross-compile a Windows executable from the same setup:

```bash
make build-windows
```

That writes the executable to `bin/win/go-fossil.exe`.

To build and run in one command:

```bash
make run
```

## Scope Guardrails

- Favor readable prototypes over premature optimization.
- Prefer one clear gameplay loop over many half-finished systems.
- Keep each change narrow enough that another agent can pick up the next step with minimal context.
- Avoid coupling terrain, rendering, and gameplay code too early.

## Current State

The repository is at project bootstrap stage. Documentation currently serves as the main source of shared context for future coding work.
