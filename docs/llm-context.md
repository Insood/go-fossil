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
- Terrain chunks are 8x8 tile render meshes derived from chunk metadata and interpolated from 9x9 inline height samples.
- Startup loads the authored default chunk at `(0,0)`.
- Generated terrain chunks are added as the running score increases, placed on exposed edges of the loaded chunk set.
- Generated chunks use random height samples that match loaded neighbor borders and random artifact placements from loaded artifact definitions.
- Artifacts are embedded as baked texture overlays plus a per-pixel artifact ID mask.
- The salvage loop is movement, downward drone-camera aiming, laser burning, cutout detection, animated fragment pickup, scoring, charging-pad drop-off, and chunk expansion.
- Successful laser terrain strikes spawn short-lived tinted cube particles that travel upward in a narrow cone and fade out.

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
9. `SoundSystem`
10. `ParticleSystem`
11. `ArtifactCutoutDetectionSystem`
12. `ArtifactFragmentPickupSystem`
13. `ArtifactFragmentDropOffSystem`
14. `ChunkSpawnerSystem`
15. `RenderSystem3D`
16. `UserInterfaceSystem`
17. `TutorialSystem`
18. `DebugRender3DSystem`
19. `DebugRenderSystem2D`

All systems implement `Initialize`, `Update`, and `Unload`. Initialization and updates follow registration order; shutdown calls `Unload` in reverse order before manager and asset teardown. Most systems currently use a no-op unload, while the fragment pickup and drop-off systems release generated models that remain in flight.

Important ownership notes:

- `AssetManager` loads runtime images, textures, GIF animations, shaders, models, streamed sounds, and artifact definitions from disk beside the built executable.
- `ChunkManager` loads authored chunk JSON, generates runtime chunks, builds terrain meshes/textures, caches chunks, samples terrain height, registers terrain chunk ECS entities, spawns chunk-owned model renderables from authored placements as shadow casters and receivers, and applies burn marks.
- Authored artifact and model placement coordinates are stored in baked texture pixels and converted to world units with `terrainTexturePixelsPerTile`.
- `ArtifactManager` owns runtime artifact records, unique artifact IDs, fragment collection state, and fragment textures.
- `DroneFireControlSystem` owns the drone viewport cursor, clamps mouse motion to the viewport, hides the OS cursor during gameplay, maps mouse/gamepad aiming into normalized drone viewport coordinates from -1 to 1 on each axis, and stores current and previous cursor/firing state on `DroneFireControl`. It shows the OS cursor and clears firing state while debug overlays are visible so raygui controls can be clicked.
- `LaserSystem` maps the stored normalized drone viewport cursor to terrain-sampled world targets, interpolates between consecutive firing cursors at the configured pixel step, stamps the chunk burn overlay while firing, drains drone battery charge once per active firing update, and marks damaged chunks for cutout scanning. Lasers only fire while battery charge is positive.
- `LaserSystem` also creates particle entities at successful terrain burn positions. These particles reuse the shared cube model, do not cast or receive shadows, and have no gameplay interaction.
- `SoundSystem` tracks every loaded sound stream by name and plays the `burning` stream while any laser is active.
- `ParticleSystem` advances particle lifetimes, fades `Renderable.tint`, and removes expired particle entities. `PhysicsSystem` moves particles because they carry `Velocity3`.
- `ArtifactCutoutDetectionSystem` periodically scans damaged chunks, flood-fills remaining artifact ID regions, accepts regions below `MaximumRegionSize`, scores recovered artifact pixels, creates fragments for regions at or above `artifactFragmentMinPixels`, spawns pickup planes at cutout centers, clears accepted artifact overlay pixels, and softens the burn overlay so the terrain shader renders a shallow divot.
- `ArtifactFragmentPickupSystem` lifts each fragment plane for 0.35 seconds, then eases it toward the live drone position for 0.65 seconds while tilting and shrinking it. Arrival marks the fragment collected, awards its score, exposes it to the inventory UI and tutorial, unloads the generated plane model, and removes the pickup entity.
- `ArtifactFragmentDropOffSystem` starts a deposit when the drone moves within 0.5 X/Z units of the authored charging pad. It launches collected fragments oldest-first from the drone at 0.25-second intervals, hides each launched fragment from inventory, and moves its full-size plane at a constant speed into the pad before unloading the plane model.
- `ChunkSpawnerSystem` watches collected score deltas and adds generated chunks after enough artifact value is recovered.
- `TutorialSystem` owns the active tutorial step, starts each run at tutorial step 1, tracks the drone's starting X/Z position, advances to step 2 once the drone moves on X or Z, spawns red shader-styled tutorial cones over artifact centers for step 2, advances to step 3 once the drone moves within 0.5 X/Z units of a cone, removes tutorial marker entities, advances to step 4 once the drone viewport cursor moves at least 25% of the drone viewport, advances to step 5 once any laser is active, advances to step 6 once the artifact manager has at least one collected fragment, spawns a tutorial cone above the authored charging pad, and completes step 6 once the drone is within `tutorialArtifactMarkerProximity` on X/Z of the pad.
- `RenderSystem3D` owns the shadow pass, main scene pass, drone bottom-camera viewport pass, laser rendering, slope shade shader tuning, and shadow-depth export.
- `UserInterfaceSystem` draws total score, the labelled drone battery bar, the drone viewport, the reticle, and recent collected fragments that have not yet been dropped off, with weight and score.
- `DebugRender3DSystem` draws axes, the drone ground ray, the light guide, and artifact ID labels when debug overlays are visible.
- `DebugRenderSystem2D` owns the raygui shadow tuning overlay, including the normal-based terrain slope shade strength.

## Rendering Notes

- `Framebuffer` in `cmd/game/framebuffer.go` wraps off-screen render targets.
- The scene has one `Light` entity. `LightSystem` rebuilds its orthographic camera from component data every frame.
- `Renderable` carries `castsShadow` and `receivesShadow` flags for render-pass participation.
- Terrain uses the dedicated `terrain_shader`, with terrain-only slope, artifact, burn, and cutout logic. Shadow-receiving non-terrain models use `model_shadow_receiver`, which preserves model albedo textures and applies the same shadow-map projection without terrain-specific code. Model receiver shadow depth is bound to every material through `MapNormal`/`texture2`, including raylib's extra default GLB material; terrain shadow depth stays on `MapHeight`/`shadowMap`.
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
