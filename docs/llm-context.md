# LLM Context

This repo is a small 3D salvage game built in Go and raylib. If you are a coding agent looking for where to make a change, start here.

## Start Here

Read these files first:

- [README.md](../README.md) for the project summary and repo shape
- [docs/technical-direction.md](technical-direction.md) for architecture and ownership guidance

## Code Shape

- `cmd/game` is the executable game and owns most gameplay code.
- `cmd/game/components.go` defines ECS components.
- `cmd/game/*_system.go` contains one system per file.
- `cmd/game/config.go` stores gameplay constants and tuning values.
- `cmd/game/game.go` owns initialization, system order, and asset lifecycle.
- `internal/terrain` owns terrain chunk parsing, validation, mesh data, baked texture generation, and terrain sampling helpers.

## Current Game

- The game uses an ECS architecture with `mlange-42/ark`.
- The world is 3D and rendered with an orthographic main camera for an isometric-style presentation.
- The player controls one drone with keyboard or gamepad movement.
- The drone has a battery charge that gates laser firing and is shown above the drone viewport.
- The drone follows loaded terrain height with a constant hover offset and a small sine-wave hover motion.
- Drone movement is clamped so its 1x1 footprint stays inside loaded terrain chunk extents in X/Z.
- The main camera uses a sliding dead zone so it only follows the drone once the drone moves far enough from center.
- A single shadow-casting light tracks the drone overhead so the shadow map stays centered on the active area.
- Terrain chunks are 8x8 tile meshes derived from chunk metadata and 9x9 inline height samples.
- Startup loads the authored default chunk at `(0,0)`.
- Generated terrain chunks are added as the running score increases, placed on exposed edges of the loaded chunk set.
- Generated chunks use random height samples that match loaded neighbor borders and random artifact placements from loaded artifact definitions.
- Artifacts are embedded as baked texture overlays plus a per-pixel artifact ID mask.
- The salvage loop is movement, downward drone-camera aiming, laser burning, cutout detection, fragment scoring, and chunk expansion.

## Current Systems

Systems are registered in this order in `cmd/game/game.go`:

1. `InputSystem`
2. `DroneInputSystem`
3. `DroneFireControlSystem`
4. `PhysicsSystem`
5. `DroneHeightSystem`
6. `CameraSystem`
7. `LightSystem`
8. `LaserSystem`
9. `ArtifactCutoutDetectionSystem`
10. `ChunkSpawnerSystem`
11. `RenderSystem3D`
12. `UserInterfaceSystem`
13. `DebugRender3DSystem`
14. `DebugRenderSystem2D`

Important ownership notes:

- `AssetManager` loads runtime images, textures, shaders, models, and artifact definitions from disk beside the built executable.
- `ChunkManager` loads authored chunk JSON, generates runtime chunks, builds terrain meshes/textures, caches chunks, samples terrain height, registers terrain chunk ECS entities, and applies burn marks.
- `ArtifactManager` owns runtime artifact records, unique artifact IDs, scored fragment records, and fragment textures.
- `DroneFireControlSystem` owns the drone viewport cursor, clamps mouse motion to the viewport, hides the OS cursor during gameplay, maps gamepad target axes into viewport space, and stores current and previous cursor/firing state on `DroneFireControl`. It shows the OS cursor and clears firing state while debug overlays are visible so raygui controls can be clicked.
- `LaserSystem` maps the stored drone viewport cursor to terrain-sampled world targets, interpolates between consecutive firing cursors at the configured pixel step, stamps the chunk burn overlay while firing, drains drone battery charge once per active firing update, and marks damaged chunks for cutout scanning. Lasers only fire while battery charge is positive.
- `ArtifactCutoutDetectionSystem` periodically scans damaged chunks, flood-fills remaining artifact ID regions, accepts regions below `MaximumRegionSize`, scores recovered artifact pixels, creates fragments for regions at or above `artifactFragmentMinPixels`, clears accepted artifact overlay pixels, and softens the burn overlay.
- `ChunkSpawnerSystem` watches score deltas and adds generated chunks after enough artifact value is recovered.
- `RenderSystem3D` owns the shadow pass, main scene pass, drone bottom-camera viewport pass, laser rendering, slope shade shader tuning, and shadow-depth export.
- `UserInterfaceSystem` draws total score, the drone battery bar, the drone viewport, the reticle, and recent fragment thumbnails with weight and score.
- `DebugRender3DSystem` draws axes, the drone ground ray, the light guide, and artifact ID labels when debug overlays are visible.
- `DebugRenderSystem2D` owns the raygui shadow tuning overlay, including the normal-based terrain slope shade strength.

## Rendering Notes

- `Framebuffer` in `cmd/game/framebuffer.go` wraps off-screen render targets.
- The scene has one `Light` entity. `LightSystem` rebuilds its orthographic camera from component data every frame.
- `Renderable` carries `castsShadow` and `receivesShadow` flags for render-pass participation.
- Terrain models use the `shadow_receiver` shader, with albedo, artifact emission, burn occlusion, and shadow depth textures bound through material maps. The shader also darkens sloped terrain based on surface normals and `slopeShadeStrength`.
- `F10` toggles debug overlays. The left gamepad trigger also toggles debug overlays.
- `F11` exports the shadow depth framebuffer for inspection.

## Change Map

| Feature type | Usually change these files |
| --- | --- |
| New ECS data | `cmd/game/components.go` |
| New gameplay system | `cmd/game/*_system.go` |
| New render behavior | `cmd/game/render_system.go` or a dedicated render system |
| Framebuffer / render-target setup | `cmd/game/framebuffer.go` |
| New input behavior | `cmd/game/input_system.go`, `cmd/game/drone_input_system.go`, or `cmd/game/drone_fire_control_system.go` |
| System order changes | `cmd/game/game.go` |
| Gameplay constants | `cmd/game/config.go` |
| Shared small helpers | `cmd/game/utils.go` |
| Terrain loading, generation, or sampling rules | `internal/terrain/*` and `cmd/game/chunk_manager.go` |
| Artifact placement, masking, scoring, or fragments | `cmd/game/artifact*.go` |

## Implementation Pattern

When adding a feature, prefer this order:

1. Add the smallest code needed to make the behavior visible or testable.
2. Put the rule in the narrowest file that clearly owns it.
3. Add or update ECS components and systems.
4. Wire the system into `cmd/game/game.go` in the correct order.
5. Update existing docs that describe changed behavior.
6. Add focused new docs when the changed behavior has no natural home in the current docs, and link them from `README.md` or this file.

Documentation is required for changes that affect ownership, coordinate assumptions, camera behavior, terrain behavior, artifact behavior, scoring, controls, rendering behavior, system order, or gameplay rules. Keep docs in present tense and describe only the implemented behavior.

## Working Style

- Make very small changes.
- Prefer vertical slices over broad scaffolding.
- Leave the project runnable after each meaningful step when possible.
- Avoid slices that only create abstractions with no immediate behavior unless they unblock a concrete change.
- Prefer visible progress, low coupling, reversible decisions, simple code, and documented assumptions.
- In final notes, include what changed, how the result was verified, and any tracked files intentionally left untouched.

## Code Style

- Keep code direct and readable.
- Prefer the smallest implementation that clearly expresses the rule.
- Favor battle-tested libraries over custom NIH code when they materially simplify implementation.
- Prefer dependencies already present in `go.mod` before adding anything new.
- Ask for permission before adding a new dependency.
- Avoid redundant defensive checks when the surrounding code or library already guarantees the invariant.
- Let hard programmer errors panic instead of hiding them behind silent branches or broad fallback handling.
- Do not add extra abstractions, interfaces, or helper layers unless they clearly reduce duplication or clarify ownership.

## Things Not To Do

- Do not move core gameplay logic into `main.go`.
- Do not create new packages without a clear ownership reason.
- Do not introduce speculative systems for inventory, progression, enemies, hazards, or AI.
- Do not replace the current texture/mask-based cutting loop with a larger terrain architecture unless the task explicitly calls for that change.
- Do not leave architectural decisions undocumented if other agents depend on them.
- Do not leave stale documentation behind when code behavior changes.
