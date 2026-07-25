# LLM Context

This repo is a small 3D salvage game built in Go and raylib. If you are a coding agent looking for where to make a change, start here.

## Start Here

Read these files first:

- [README.md](../README.md) for the player-facing project summary and build instructions
- [docs/technical-direction.md](technical-direction.md) for architecture and ownership guidance

## Code Shape

- `cmd/game` is the executable game and owns most gameplay code.
- `cmd/game/splash_*.go` defines the ECS-backed generated terrain scene shown before gameplay initialization.
- `cmd/game/components.go` defines ECS components.
- `cmd/game/*_system.go` contains one system per file.
- `cmd/game/config.go` stores gameplay constants and tuning values.
- `cmd/game/main.go` owns the shared `AssetManager` lifecycle; `cmd/game/game.go` owns gameplay initialization, system order, and scene-local teardown.
- `internal/terrain` owns terrain chunk parsing, validation, mesh data, baked texture generation, and terrain sampling helpers.

## Current Game

- The game uses an ECS architecture with `mlange-42/ark`.
- Application startup first runs a splash scene with its own Ark world, nine randomly generated chunks centered on `(0,0)`, a simulated drone, fixed gameplay camera framing, and splash text. The drone chooses a random full-speed X/Z heading once per second, uses normal physics and terrain-bound clamping, and stays within 6 world units on each axis of a point offset 2 world units toward positive X/Z from the center chunk's geometric center. It also alternates between firing and cooldown periods lasting 0.2–0.6 seconds, steering its terrain target with a bounded random-walk offset. It cannot be controlled by the player. Pressing Space or gamepad A unloads the splash runtime scene and initializes the gameplay world.
- The world is 3D and rendered with an orthographic main camera for an isometric-style presentation.
- The player controls one drone with keyboard or gamepad movement.
- The drone has a battery charge that constantly drains, gates laser firing, and is shown above the drone viewport. Flight drain increases with carried cargo weight and while moving.
- The drone follows loaded terrain height with a constant hover offset and a small sine-wave hover motion.
- Battery depletion removes player control and hover motion, then the drone sinks until generic terrain collision settles it at sampled terrain height.
- Drone movement is clamped so its 1x1 footprint stays inside loaded terrain chunk extents in X/Z.
- The main camera uses a sliding dead zone so it only follows the drone once the drone moves far enough from center.
- A single shadow-casting light tracks the drone overhead so the shadow map stays centered on the active area.
- Terrain chunks are 8x8 tile render meshes derived from chunk metadata and interpolated from 9x9 inline height samples.
- Startup loads the authored default chunk at `(0,0)`.
- Generated terrain chunks are added as the running score increases, placed on exposed edges of the loaded chunk set.
- Generated chunks use random height samples that match loaded neighbor borders and independent random artifact placements weighted by each loaded definition's positive `relative_scarcity`; artifact types may repeat within a chunk.
- Artifacts are embedded as baked texture overlays plus a per-pixel artifact ID mask.
- The salvage loop is movement, downward drone-camera aiming, laser burning, cutout detection, fragment rising and proximity pickup within drone cargo capacity, charging-pad drop-off and scoring, and chunk expansion.
- Successful laser terrain strikes spawn short-lived tinted cube particles that travel upward in a narrow cone and fade out.

## Current Systems

Systems are registered in this order in `cmd/game/game.go`:

1. `InputSystem`
2. `DroneInputSystem`
3. `BatteryDrainSystem`
4. `GameOverDetectionSystem`
5. `PlayerDroneFireControlSystem`
6. `PhysicsSystem`
7. `DroneHeightSystem`
8. `TerrainCollisionDetectionSystem`
9. `CameraSystem`
10. `LightSystem`
11. `PlayerDroneFireTargetSystem`
12. `LaserSystem`
13. `SoundSystem`
14. `ParticleSystem`
15. `ArtifactCutoutDetectionSystem`
16. `MovementAnimationSystem`
17. `ArtifactFragmentPickupSystem`
18. `ArtifactFragmentDropOffSystem`
19. `ChunkSpawnerSystem`
20. `RenderSystem3D`
21. `UserInterfaceSystem`
22. `TutorialSystem`
23. `DebugRender3DSystem`
24. `DebugRenderSystem2D`

All systems implement `Initialize`, `Update`, and `Unload`. The main loop snapshots raylib's frame time into `Game.FrameTime` before each system update pass, so systems use one consistent delta per frame. Initialization and updates follow registration order; shutdown calls `Unload` in reverse order before manager and asset teardown. Most systems currently use a no-op unload, while the fragment pickup system releases all generated fragment models still in the world and the drop-off system releases generated models that remain in flight.

Before these gameplay systems are created, the splash runs `SplashScreenDroneControlSystem`, `PhysicsSystem`, `DroneHeightSystem`, `LightSystem`, `SplashScreenDroneFireTargetSystem`, `LaserSystem`, `ParticleSystem`, and `RenderSystem3D` against its temporary world, followed by its own start-input and text-render systems. The movement controller changes the drone's random velocity once per second and clamps it both to loaded terrain and to a configured square with a 6-world-unit X/Z radius around the offset splash focus point. The fire-target system random-walks an X/Z offset within `[-0.5, 0.5]`, alternates random 0.2–0.6-second firing and cooldown states, and emits terrain-height world targets while firing. The splash drone has a 1,000,000-charge battery; its beam burns terrain and produces particles, while sound, cutout detection, and fragment gameplay remain disabled. The 3D renderer omits the drone viewport pass. Splash systems initialize in registration order and unload in reverse order; generated chunks, their GPU resources, the shadow framebuffer, managers, entities, and world are destroyed before gameplay loads the authored default chunk. After game over, Space or gamepad A unloads gameplay and creates a fresh splash scene; starting again creates a new gameplay world and zeroed score. `main` owns the disk-loaded `AssetManager`, lends it to each scene, and unloads it only when the application exits.

Important ownership notes:

- `AssetManager` loads runtime images, textures, GIF animations, shaders, models, streamed sounds, and artifact definitions from disk beside the built executable, and totals artifact `relative_scarcity` values for generated-placement weighting.
- `ChunkManager` loads authored chunk JSON, generates runtime chunks, builds terrain meshes/textures, caches chunks, samples terrain height, registers terrain chunk ECS entities, resolves typed entity placements through game-owned factories, tracks those entities for chunk teardown, and applies burn marks.
- Authored artifact and entity placement coordinates are stored in baked texture pixels and converted to world units with `terrainTexturePixelsPerTile`. Entity placement types are gameplay archetypes; their factories own model selection, render settings, and ECS component composition.
- `ArtifactManager` owns runtime artifact records, unique artifact IDs, fragment collection state, fragment textures, and carried-fragment weight calculation.
- `BatteryDrainSystem` applies time-based drain after drone input. Its rate is 0.25 charge per second plus carried-weight percentage multiplied by the 1.0 cargo modifier, reaching 1.25 charge per second at the 12,000-unit limit; movement does not alter drain. Completed drop-offs accumulate `score * 0.05` in a drone `BatteryRecharge` component. The battery system transfers up to 1 charge from that reservoir per update, removes it when exhausted, clamps battery charge to 100, and discards pending recharge when a transfer would exceed 100. Drained charge is clamped at zero.
- `GameOverDetectionSystem` finds the `PlayerControlled` drone at zero battery, zeros X/Z velocity, applies the configured negative Y velocity, removes `PlayerControlled` and `HoverMotion`, and creates one standalone `GameOver` marker. This is a terminal state even if pending drop-off activity later adds battery charge. `TerrainCollisionDetectionSystem` checks entities with non-zero Y velocity after physics and drone height updates, snapping any below-terrain Y position to sampled terrain and zeroing its Y velocity.
- `PlayerControlled` gates `DroneInputSystem`, `PlayerDroneFireControlSystem`, and `PlayerDroneFireTargetSystem`; the automated splash drone does not carry it. The fire-control system owns the player drone viewport cursor, clamps mouse motion to the viewport, hides the OS cursor during gameplay, maps mouse/gamepad aiming into normalized coordinates from -1 to 1 on each axis, and stores current and previous cursor/firing state on `PlayerFireInput`. It shows the OS cursor and clears firing state while debug overlays are visible so raygui controls can be clicked.
- `PlayerDroneFireTargetSystem` clears the controlled player's per-frame `DroneFireTargets`, interpolates between consecutive firing cursors at the configured pixel step, and converts valid normalized cursors into terrain-sampled world coordinates after drone movement and height updates.
- `LaserSystem` consumes world-space `DroneFireTargets` without depending on player input. With positive battery and at least one target, it presents the final target, drains charge once, stamps every target into the chunk burn overlay, and marks damaged chunks for cutout scanning. It clears target buffers whether it fires or the battery is empty.
- `LaserSystem` also creates particle entities at successful terrain burn positions. These particles reuse the shared cube model, do not cast or receive shadows, and have no gameplay interaction.
- `SoundSystem` tracks every loaded sound stream by name and plays the looping `burning` stream while any laser is active. It also consumes short-lived `SoundPlaybackRequest` ECS entities, playing non-burning streams once and retaining repeated requests until earlier playback of the same stream finishes.
- `ParticleSystem` advances particle lifetimes, fades `Renderable.tint`, and removes expired particle entities. `PhysicsSystem` moves particles because they carry `Velocity3`.
- `ArtifactCutoutDetectionSystem` periodically scans damaged chunks, flood-fills remaining artifact ID regions, accepts regions below `MaximumRegionSize`, scores recovered artifact pixels, records each fragment's 0-to-1 recovered-pixel grade (capped at full recovery), creates fragments for regions at or above `artifactFragmentMinPixels`, spawns textured fragment planes at cutout centers, queues one `pop.wav` playback request per spawned fragment, clears accepted artifact overlay pixels, and softens the burn overlay so the terrain shader renders a shallow divot.
- `MovementAnimationSystem` owns reusable translation animations. Each movement stores start and fallback target positions, duration, elapsed time, and linear or cubic easing; it can optionally resolve a live entity target such as the drone. Completion snaps to the target and removes only the movement component, allowing gameplay systems to interpret the finished stage.
- New fragment planes receive an ease-out movement that lifts them for 0.35 seconds. Once the movement component is removed, the fragment remains hovering at the raised position and is ready for pickup.
- `ArtifactFragmentPickupSystem` gives the drone a 12,000-unit carry limit. Each update it may start homing the nearest ready fragment within 0.5 X/Z units whose full pixel weight fits, using fragment ID to break distance ties. It starts at most one pickup per update, and homing fragments reserve capacity immediately. Pickup adds an ease-in movement targeting the live drone for 0.65 seconds while the pickup system applies camera-facing tilt and shrink presentation. When that movement ends, pickup marks the fragment collected, exposes it to the inventory UI and tutorial, unloads the generated plane model, and removes the world entity without changing total score.
- `ArtifactFragmentDropOffSystem` tracks one ejection cooldown and, whenever it is ready, rechecks current cargo and whether the drone is within 0.5 X/Z units of the nearest pad. It launches the oldest available fragment, resets the 0.25-second cooldown, hides the fragment from inventory, and gives its plane a linear movement whose distance-derived duration preserves the configured constant speed. It retains no drop-off queue, so leaving the pad stops further ejections. When a plane reaches the pad, the system adds its fragment score to `Game.TotalScore`, adds `score * 0.05` to the drone's pending battery recharge, queues one `score.wav` playback request, unloads the model, and removes the entity.
- `ChunkSpawnerSystem` watches deposited score deltas and adds generated chunks after enough artifact value is delivered.
- `TutorialSystem` owns the active tutorial step, starts each run at tutorial step 1, tracks the drone's starting X/Z position, advances to step 2 once the drone moves on X or Z, spawns red shader-styled tutorial cones over artifact centers for step 2, advances to step 3 once the drone moves within 0.5 X/Z units of a cone, removes tutorial marker entities, advances to step 4 once the drone viewport cursor moves at least 25% of the drone viewport, advances to step 5 once any laser is active, advances to step 6 once the artifact manager has at least one collected fragment, queries the nearest `ChargingPad` entity to place the return-home cone, advances to step 7 once the drone is within `tutorialArtifactMarkerProximity` on X/Z of the pad, displays "Collect more!" for five seconds, and then completes. It pauses updates and drawing while `GameOver` exists.
- `RenderSystem3D` owns the shadow pass, main scene pass, drone bottom-camera viewport pass, laser rendering, slope shade shader tuning, and shadow-depth export.
- `UserInterfaceSystem` draws total score, aligned labelled cargo and battery bars above the drone viewport, the drone viewport, the reticle, and recent collected fragments that have not yet been dropped off, with score and a color-coded `[SUPER]`/`[A]`/`[B]`/`[C]`/`[F]` grade. The gold cargo bar shows collected, non-dropped fragment weight against the 12,000-unit capacity; homing fragments are not included. The battery bar is lime at 20% charge or higher and red below 20%. While `GameOver` exists it replaces that HUD with centered `GAME OVER` and final-score text over the still-rendered 3D scene.
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
| New input behavior | `cmd/game/input_system.go`, `cmd/game/drone_input_system.go`, or `cmd/game/player_drone_fire_control_system.go` |
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
