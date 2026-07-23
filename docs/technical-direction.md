# Technical Direction

## Stack

- Language: Go
- Rendering and platform layer: raylib
- Architecture: ECS via `mlange-42/ark`
- Image processing: `github.com/disintegration/gift` for artifact sprite rotation and terrain tile scaling

## Rendering Model

The game is a 3D world presented through an orthographic main camera for an isometric-style view. Gameplay positions remain fully 3D even when movement is constrained to X/Z over terrain.

Current rendering passes:

- a shadow pass into `Game.shadowFramebuffer` from the active `Light` camera
- a main scene pass through `Game.camera`
- a downward drone-camera pass into `Game.droneFramebuffer`
- 2D UI and debug overlay passes after the 3D scene

The main camera keeps a fixed offset from its focus point and slides only when the drone exits the configured X/Z or Y dead zone. The light camera is orthographic and tracks the drone overhead every frame.

Terrain models use the `shadow_receiver` shader. Terrain albedo, artifact overlay, burn overlay, and shadow depth textures are bound through raylib material maps. The shader applies shadow-map darkening and a light normal-based slope shade: flat upward-facing terrain keeps its original color, while sloped terrain is darkened by `slopeShadeStrength` according to the surface normal's Y component. Accepted cutout regions are visually lowered by `terrainCutoutDivotDepth` using the burn overlay as a shader displacement mask. Renderable ECS models use `castsShadow` and `receivesShadow` flags to participate in the shadow pass and shadow receiver setup.

## Terrain Model

Terrain chunks use a fixed 8x8 tile footprint with a 9x9 height sample grid. Terrain tile images are scaled into the baked base texture with `gift.Resize` using nearest-neighbor resampling.

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

- a raylib render mesh/model generated from interpolated height samples
- a baked base albedo image/texture assembled from tile textures
- a baked artifact overlay image/texture
- a per-pixel artifact ID mask
- a burn overlay image/texture
- registered runtime artifact records
- a `TerrainChunkComponent` ECS entity

`internal/terrain` uses raylib vector helpers for mesh normal accumulation while still owning the terrain mesh data shape.

Laser cutting is currently represented by texture and mask state. Burn marks paint the burn overlay and clear artifact IDs at the affected pixels. Accepted cutout regions soften the burn overlay and are rendered as shallow shader-side divots. Height samples are not mutated by cutting.

## Artifact And Salvage Model

Artifact definitions live in `cmd/game/assets/artifacts/*.json` and reference texture images. Streamed sounds live in `cmd/game/assets/sounds` and are loaded through raylib music streams. `AssetManager` loads definitions and verifies referenced images. Artifact placement rotation uses `gift.Rotate` with nearest-neighbor interpolation, and each runtime `Artifact.Size` is counted from the rotated placement image.

Chunk artifact placements are baked into two layers:

- `ArtifactImage`, the visible overlay texture composited with standard `image/draw`
- `ArtifactData`, the per-pixel artifact ID mask used for cutout detection and scoring

`ArtifactCutoutDetectionSystem` scans damaged chunks every `artifactCutoutDetectionScanTicks`. It flood-fills remaining artifact ID regions, accepts regions smaller than `MaximumRegionSize`, scores the recovered pixels by artifact value and the rotated runtime artifact size, creates in-memory fragments for regions at or above `artifactFragmentMinPixels`, clears the accepted overlay pixels, softens the burn overlay, and adds created fragment scores to `Game.TotalScore`.

## Entity Model

The current runtime entities are:

- one player drone with position, velocity, renderable model, drone tag, hover motion, laser, fire-control, and battery components
- one light entity with a mutable orthographic camera
- one terrain chunk entity per loaded chunk
- short-lived laser strike particle entities with position, velocity, renderable cube model, and particle lifetime state

Current ECS components:

- `Position3`
- `Velocity3`
- `Renderable`
- `HoverMotion`
- `Light`
- `Laser`
- `Particle`
- `TerrainChunkComponent`
- `TerrainChunkDamaged`
- `Drone`
- `Battery`
- `DroneFireControl`

## System Ownership

- Input: gamepad quit handling in `InputSystem`; movement in `DroneInputSystem`; aim/firing state in `DroneFireControlSystem`. Drone fire-control cursors are stored as normalized drone viewport coordinates from -1 to 1 on each axis. Debug overlays release drone fire control so the OS cursor can interact with raygui controls.
- Motion: `PhysicsSystem` applies velocity; `DroneHeightSystem` snaps the drone to terrain height plus hover offset.
- Presentation: `CameraSystem` updates the main orthographic camera; `LightSystem` updates the shadow camera.
- Salvage: `LaserSystem` maps the drone viewport cursor path to terrain, applies burns while battery charge is positive, drains charge once per active firing update, and spawns short-lived cube particles at successful terrain strikes; `ArtifactCutoutDetectionSystem` scores accepted artifact regions.
- Sound: `SoundSystem` tracks loaded sound streams by name and updates raylib music streams while their gameplay-driven sound state is playing. The burning stream plays while any laser is active.
- Particles: `ParticleSystem` advances particle lifetimes, fades renderable tint alpha, and removes expired particle entities. Particle motion uses the shared velocity-driven physics system.
- World growth: `ChunkSpawnerSystem` adds generated chunks after score increases.
- Tutorial: `TutorialSystem` owns the active tutorial step, starts each run at step 1, advances to step 2 once the drone moves away from its starting X/Z position, spawns red shader-styled tutorial cones over artifact centers for step 2, advances to step 3 once the drone moves within 0.5 X/Z units of a cone, removes tutorial marker entities, completes step 3 once the drone viewport cursor moves at least 25% of the drone viewport, and draws the active tutorial prompt.
- Rendering: `RenderSystem3D` owns shadow, scene, drone viewport, laser rendering, and depth export.
- UI/debug: `UserInterfaceSystem`, `TutorialSystem`, `DebugRender3DSystem`, and `DebugRenderSystem2D` draw score, the drone battery bar, fragment thumbnails, viewport, reticle, the active tutorial prompt, artifact labels, debug guides, shadow tuning controls, and the slope shade tuning control.

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
