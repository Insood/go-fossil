# LLM Context

This repo is a small 3D salvage game built in Go and raylib. If you are a coding agent looking for where to make a change, start here.

## Start Here

Read these files first:

- [README.md](../README.md) for the project summary and repo shape
- [docs/project-brief.md](project-brief.md) for the gameplay pillars and prototype goals
- [docs/technical-direction.md](technical-direction.md) for architecture and ownership guidance
- [docs/agent-guide.md](agent-guide.md) for small-slice workflow rules

## Default Code Shape

This project is intended to follow the same overall structure as `go-towerdefense`.

- `cmd/game` is the executable game
- `cmd/game/components.go` defines ECS components
- `cmd/game/*_system.go` contains one system per file
- `cmd/game/config.go` stores gameplay constants
- `cmd/game/game.go` owns initialization, system order, and asset lifecycle
- `internal/*` packages are reserved for focused game-rule domains that need separation

## What This Repo Is

- The game uses an ECS architecture with `mlange-42/ark`.
- The world is 3D and rendered with an orthographic camera to feel isometric.
- Terrain starts as a mesh derived from chunk metadata and inline height samples.
- Each authored chunk currently maps to one 8x8 terrain chunk with a 9x9 height sample grid.
- The core loop is movement, discovery, cutting, and pickup under light pressure.
- Gameplay code should stay small and easy to follow.

## Early Invariants

- World coordinates should remain truly 3D.
- Camera choices are presentation, not world-model rules.
- Terrain should be easy to render and easy to mutate experimentally.
- The cutting mechanic is more important than early generality.
- A visible, playable slice is better than a broad abstraction pass.

## Where To Make Changes

Use this as the default change map:

| Feature type | Usually change these files |
| --- | --- |
| New ECS data | `cmd/game/components.go` |
| New gameplay system | `cmd/game/*_system.go` |
| New render behavior | `cmd/game/render_system.go` or a dedicated render system |
| Framebuffer / render-target setup | `cmd/game/framebuffer.go` |
| New input behavior | `cmd/game/input_system.go` |
| System order changes | `cmd/game/game.go` |
| Gameplay constants | `cmd/game/config.go` |
| Shared small helpers | `cmd/game/utils.go` |
| Terrain loading, generation, or mutation rules | `internal/terrain/*` |
| Salvage/extraction rules that outgrow `cmd/game` | `internal/salvage/*` |

## Implementation Pattern

When adding a feature, prefer this order:

1. Add the smallest code needed to make the behavior visible or testable.
2. Put the rule in the narrowest file that clearly owns it.
3. Add or update ECS components and systems.
4. Wire the system into `cmd/game/game.go` in the correct order.
5. Update docs if the change affects ownership, coordinate assumptions, camera behavior, terrain behavior, or system order.

## Current Implementation Notes

- `AssetManager` owns loading runtime assets from disk beside the built executable. `ChunkManager` owns loading terrain chunk JSON from disk, converting it into runtime terrain chunks, and caching the built meshes/textures. `internal/terrain` owns chunk JSON parsing, validation, terrain mesh data generation, baked terrain image composition, and world-to-terrain UV helpers.
- There is a dedicated `Framebuffer` wrapper in `cmd/game/framebuffer.go` for off-screen render targets.
- The scene currently has a single `Light` entity. `LightSystem` rebuilds its camera from component data, and the render pipeline consumes that camera.
- `Renderable` carries `castsShadow` and `receivesShadow` flags so future render passes can filter participation without adding extra ECS components.
- `LaserSystem` maps held left-click input inside the drone viewport to a terrain-sampled target point for the player drone's `Laser` component.
- `RenderSystem3D` currently owns the main scene render flow, the drone bottom-camera viewport pass, and the temporary shadow-depth debug pass.
- `DebugRenderSystem2D` owns the top-right raygui overlay for live shadow tuning controls, including the current light origin and orthographic size.
- `F10` toggles the debug overlays, and `F11` exports the framebuffer depth texture for inspection.
- The current shadow work is intentionally incremental; prefer tiny, verifiable slices before introducing actual shadow sampling into the lit scene.

## Things Not To Do

- Do not move core gameplay logic into `main.go`.
- Do not create new packages without a clear ownership reason.
- Do not introduce speculative systems for inventory, progression, or AI depth.
- Do not make terrain architecture complex before the cutting mechanic proves itself.
- Do not leave architectural decisions undocumented if future agents will depend on them.
