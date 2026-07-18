# Technical Direction

## Stack

- Language: Go
- Rendering and platform layer: raylib
- Architecture: ECS via `mlange-42/ark`
- Image processing: `github.com/disintegration/gift` for artifact sprite rotation

## Rendering Model

The game is a 3D world presented through an orthographic main camera for an isometric-style view. Gameplay positions remain fully 3D even when movement is constrained to X/Z over terrain.

Current rendering passes:

- a shadow pass into `Game.shadowFramebuffer` from the active `Light` camera
- a main scene pass through `Game.camera`
- a downward drone-camera pass into `Game.droneFramebuffer`
- 2D UI and debug overlay passes after the 3D scene

The main camera keeps a fixed offset from its focus point and slides only when the drone exits the configured X/Z or Y dead zone. The light camera is orthographic and tracks the drone overhead every frame.

Terrain models use the `shadow_receiver` shader. Terrain albedo, artifact overlay, burn overlay, and shadow depth textures are bound through raylib material maps. Renderable ECS models use `castsShadow` and `receivesShadow` flags to participate in the shadow pass and shadow receiver setup.

## Terrain Model

Terrain chunks use a fixed 8x8 tile footprint with a 9x9 height sample grid.

Authored chunks:

- live in `cmd/game/assets/terrain_chunks/*.json`
- define a chunk name, tile index grid, tile texture definitions, height samples, and artifact placements
- are parsed and validated by `internal/terrain`

Generated chunks:

- are created by `ChunkGenerator`
- start with the ground grid tile definition
- receive random height samples in `ChunkManager`
- copy loaded neighbor border heights to keep seams continuous
- receive random artifact placements from loaded artifact definitions

Built terrain chunks contain:

- a raylib mesh/model generated from height samples
- a baked base albedo image/texture assembled from tile textures
- a baked artifact overlay image/texture
- a per-pixel artifact ID mask
- a burn overlay image/texture
- registered runtime artifact records
- a `TerrainChunkComponent` ECS entity

Laser cutting is currently represented by texture and mask state. Burn marks paint the burn overlay and clear artifact IDs at the affected pixels. Height samples are not mutated by cutting.

## Artifact And Salvage Model

Artifact definitions live in `cmd/game/assets/artifacts/*.json` and reference texture images. `AssetManager` loads definitions, verifies referenced images, and computes each artifact's non-transparent pixel size. Artifact placement rotation uses `gift.Rotate` with nearest-neighbor interpolation.

Chunk artifact placements are baked into two layers:

- `ArtifactImage`, the visible overlay texture
- `ArtifactData`, the per-pixel artifact ID mask used for cutout detection and scoring

`ArtifactCutoutDetectionSystem` scans damaged chunks every `artifactCutoutDetectionScanTicks`. It flood-fills remaining artifact ID regions, accepts regions smaller than `MaximumRegionSize`, scores the recovered pixels by artifact value and original artifact size, creates in-memory fragments for regions at or above `artifactFragmentMinPixels`, clears the accepted overlay pixels, softens the burn overlay, and adds created fragment scores to `Game.TotalScore`.

## Entity Model

The current runtime entities are:

- one player drone with position, velocity, renderable model, drone tag, hover motion, laser, and fire-control components
- one light entity with a mutable orthographic camera
- one terrain chunk entity per loaded chunk

Current ECS components:

- `Position3`
- `Velocity3`
- `Renderable`
- `HoverMotion`
- `Light`
- `Laser`
- `TerrainChunkComponent`
- `TerrainChunkDamaged`
- `Drone`
- `DroneFireControl`

## System Ownership

- Input: gamepad quit handling in `InputSystem`; movement in `DroneInputSystem`; aim/firing state in `DroneFireControlSystem`.
- Motion: `PhysicsSystem` applies velocity; `DroneHeightSystem` snaps the drone to terrain height plus hover offset.
- Presentation: `CameraSystem` updates the main orthographic camera; `LightSystem` updates the shadow camera.
- Salvage: `LaserSystem` maps the drone viewport cursor path to terrain and applies burns; `ArtifactCutoutDetectionSystem` scores accepted artifact regions.
- World growth: `ChunkSpawnerSystem` adds generated chunks after score increases.
- Rendering: `RenderSystem3D` owns shadow, scene, drone viewport, laser rendering, and depth export.
- UI/debug: `UserInterfaceSystem`, `DebugRender3DSystem`, and `DebugRenderSystem2D` draw score, fragment thumbnails, viewport, reticle, artifact labels, debug guides, and shadow tuning controls.

## Repository Shape

```text
cmd/game/
  main.go
  game.go
  components.go
  config.go
  *_system.go
internal/terrain/
docs/
```

Guidelines:

- keep the main executable under `cmd/game`
- keep ECS components in `cmd/game/components.go`
- keep one system per file under `cmd/game/*_system.go`
- keep `main.go` thin
- keep constants and tuning values in `cmd/game/config.go`
- keep asset loading and cleanup in `cmd/game/game.go` and `cmd/game/asset_manager.go`
- keep terrain parsing, validation, mesh generation, baked terrain texture generation, and terrain sampling helpers in `internal/terrain`
- create another `internal/*` package only when a gameplay rule or data model has a clear independent owner

## Code Style

- Prefer direct code over defensive code when the repo or library already enforces the invariant.
- Do not add extra null checks, "should never happen" branches, or fallback paths just to feel safe.
- If a state is truly invalid for the game, let it panic loudly instead of being silently swallowed.
- Keep abstractions minimal; avoid extra interfaces and helper layers unless they make the code materially clearer.

## Performance Guidance

Keep the default implementation straightforward and measurable. Optimize only around observed costs in the current data flow, such as per-frame chunk sorting, repeated material texture binding, texture upload regions, or render-pass work.
