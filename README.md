# go-fossil

`go-fossil` is a small 3D salvage game built with Go and raylib.

The player controls a drone exploring a fossilized junkyard from an orthographic, isometric-style camera view. The drone searches for valuable artifacts embedded in terrain, aims through a downward viewport, cuts overlay regions with a laser, and recovers scored fragments.

## Core Fantasy

Fly low through a dead-tech graveyard, scan the landscape, carve out buried relics, and uncover more ground through careful salvage.

## Design Tone

The setting reads as an archaeological dig site crossed with an e-waste graveyard: dead hardware, compressed layers of discarded technology, and relics from different eras buried together.

## Current Implementation

- Engine language: Go
- Rendering library: raylib
- Entity model: ECS using [`mlange-42/ark`](https://github.com/mlange-42/ark)
- World presentation: 3D scene rendered with an orthographic camera
- Camera: dead-zone following around the player drone
- Lighting: a drone-following orthographic light and shadow map
- Terrain: 8x8 chunk meshes generated from metadata and 9x9 inline height samples
- Artifacts: JSON definitions, texture overlays, per-pixel ID masks, values, and recovered fragment records
- Cutting: laser burns painted into terrain overlay state; cutting does not mutate height samples
- Scoring: accepted artifact regions create scored fragments above the configured pixel minimum and increase total score
- World growth: generated chunks spawn on exposed loaded-chunk edges after score gains

## Documentation Map

- [docs/technical-direction.md](docs/technical-direction.md): architecture choices, current systems, and technical constraints
- [docs/llm-context.md](docs/llm-context.md): implementation map and working agreements for coding agents

## Repository Shape

- `cmd/game` holds the executable game and most ECS-facing code
- `cmd/game/components.go` defines ECS components
- `cmd/game/*_system.go` keeps one gameplay, UI, debug, or render system per file
- `cmd/game/config.go` stores gameplay constants and tuning values
- `cmd/game/game.go` owns bootstrapping, system wiring, and game-owned asset lifecycle
- `cmd/game/asset_manager.go` owns asset loading details
- `internal/terrain` owns terrain chunk parsing, validation, mesh data generation, baked terrain texture generation, and terrain sampling helpers
- `docs` holds project and agent-facing documentation

Current content format:

- chunk metadata lives in `cmd/game/assets/terrain_chunks/*.json`
- artifact definitions live in `cmd/game/assets/artifacts/*.json`
- tile layout stays in chunk JSON
- vertex heights come from inline height samples sized `(width+1) x (height+1)`
- artifact placements in chunk JSON reference artifact definitions by name
- typed entity placements in chunk JSON reference game-owned archetypes such as `charging_pad`
- `AssetManager` loads runtime images, textures, shaders, models, streamed sounds, and artifact definitions from disk beside the built executable

## Build

On WSL/Ubuntu, build the game with:

```bash
make build
```

The build writes the Linux binary to `bin/linux/go-fossil`, copies `cmd/game/assets` to `bin/linux/assets`, and uses a repo-local Go build cache. The target platform output directory is rebuilt each time so removed assets do not linger in output.

To cross-compile a Windows executable from the same setup:

```bash
make build-windows
```

That writes the executable to `bin/win/go-fossil.exe` and copies the asset tree to `bin/win/assets`.

To build and run in one command:

```bash
make run
```

## Scope Guardrails

- Favor readable prototypes over premature optimization.
- Prefer one clear gameplay loop over many half-finished systems.
- Keep each change narrow enough that another agent can pick up the next step with minimal context.
- Keep terrain, rendering, input, UI, and gameplay ownership clear.

## Current State

The repository contains a playable salvage slice: drone movement, height-sampled terrain chunks, embedded artifact overlays, viewport laser cutting, cutout detection, fragment scoring, generated chunk expansion, UI, shadow rendering, and debug overlays.


## Assets
# Fossils

Remixed from
https://opengameart.org/content/fossil-undead-rpg-enemy-mod-therapsid-charset (CC-BY 3.0 Flying Tiger Comics)
Which are themselves remixed from
https://opengameart.org/content/fossil-undead-rpg-enemy-sprites (CC-BY 3.0 Stephen Challener (Redshrike))

Cellphone?
{Unknown}

# Other Images
Steamdeck by ExxiIon ("Feel free to use and change however you like, just make sure to credit me" License)
https://www.reddit.com/r/SteamDeck/comments/trztvf/my_pixel_art_submission_for_rplace_feel_free_to/

# Sounds
Burning Sound by Freesound Community
https://pixabay.com/sound-effects/film-special-effects-burning-fire-steam-87118/

# Fonts
Grixel Acme 9 by Nikos Giannakopoulos (Creative Commons Attribution-NoDerivs 2.5)
https://www.dafont.com/grixel-acme-9.font
http://creativecommons.org/licenses/by-nd/2.5/
