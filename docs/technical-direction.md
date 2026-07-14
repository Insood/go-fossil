# Technical Direction

## Stack

- Language: Go
- Rendering and platform layer: raylib
- Architecture: ECS via `mlange-42/ark`

## Rendering Model

The game is fundamentally 3D, but presented with an orthographic camera to achieve an isometric feel.

Implications:

- world positions should remain fully 3D
- gameplay logic should not assume a 2D plane even if movement is initially constrained
- camera code should be treated as presentation, not world-model logic

## Terrain Model

Initial terrain is a mesh generated from level metadata plus a grayscale heightmap image. Authored levels currently define one fixed-size terrain chunk: 8x8 tiles with a matching 9x9 heightmap.

Desired properties:

- easy to render
- easy to inspect visually
- compatible with terrain modification experiments
- simple to validate from authored content files

Current terrain rendering approach:

- one terrain mesh generated from the level height samples for a single 8x8 chunk
- one baked albedo texture composed from tile textures
- a matching overlay texture can be added later for cut marks and painted state

Open question for future implementation:

- whether terrain cutting mutates the heightmap itself, a secondary mask, or a localized voxel/submesh structure

For early slices, prefer a fake or simplified cutting representation if it helps validate the gameplay loop faster.

## Entity Model

Use ECS for gameplay entities and systems.

Likely early entities:

- player drone
- salvageable artifact
- security drone
- junkyard rat
- terrain markers or interaction probes

Likely early component categories:

- transform / position
- velocity
- render model reference
- drone control
- collider or bounds
- salvageable
- laser cutter
- AI state
- health or integrity

## Separation Guidelines

Keep these concerns separate as long as possible:

- rendering
- input
- ECS world state
- terrain generation
- terrain modification
- gameplay rules

This will make small agent-driven slices safer and easier to review.

## Repository Shape

Because this project will be structured similarly to the earlier `go-towerdefense` codebase, this layout should be treated as the default unless a strong reason emerges to change it:

```text
cmd/game/
  main.go
  game.go
  components.go
  config.go
  *_system.go
internal/terrain/
internal/world/
internal/salvage/
docs/
```

Guidelines:

- keep the main executable under `cmd/game`
- keep ECS components in `cmd/game/components.go`
- keep one system per file under `cmd/game/*_system.go`
- keep `main.go` thin
- create `internal/*` packages only when a gameplay rule or data model clearly benefits from separate ownership
- do not split code into many packages before the first playable loop exists

Likely early internal package candidates:

- `internal/terrain` for level loading, terrain generation, and mutation rules
- `internal/world` for coordinate helpers or spatial rules if they outgrow `cmd/game`
- `internal/salvage` for extraction-specific logic if it becomes distinct from rendering and ECS wiring

## Prototype Bias

Prefer the simplest version that preserves future flexibility:

- placeholder geometry over final models
- hardcoded test maps over content pipelines
- basic ECS systems over generalized frameworks

## Performance Guidance

Do not optimize first. Establish working data flow and visual correctness before investing in terrain mutation or rendering performance.

Performance work becomes important after:

1. the terrain representation is chosen
2. cutting behavior is proven fun enough to keep
3. actual bottlenecks are measured
