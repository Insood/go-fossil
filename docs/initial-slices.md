# Initial Slices

This document suggests a practical order for the first implementation passes.

## Slice 1: Application Shell

Goal:

- create a runnable Go program that opens a raylib window and cleanly exits

Why it matters:

- verifies local toolchain and library wiring
- gives every future slice a stable entry point

## Slice 2: Orthographic Camera

Goal:

- render a basic 3D test scene with an orthographic camera that feels isometric

Why it matters:

- confirms the visual language of the game early
- exposes camera and scale issues before gameplay code grows

## Slice 3: Terrain Prototype

Goal:

- render a small terrain mesh generated from a simple heightmap

Why it matters:

- establishes the world surface that most interactions depend on

## Slice 4: ECS Bootstrap

Goal:

- create the minimal ECS world and spawn a drone entity with transform-related data

Why it matters:

- validates the chosen architecture without overcommitting to system structure

## Slice 5: Drone Control

Goal:

- move the drone around the terrain with readable controls

Why it matters:

- creates the first real player interaction

## Slice 6: Salvageable Artifact

Goal:

- place at least one visible embedded object in the terrain

Why it matters:

- turns movement into exploration

## Slice 7: Cutting Prototype

Goal:

- let the drone perform a simple cut action that visibly alters terrain or a terrain-adjacent mask

Why it matters:

- tests the game’s defining mechanic as early as possible

## Slice 8: Pickup Loop

Goal:

- allow an exposed artifact to be collected

Why it matters:

- completes the first end-to-end gameplay loop

## Slice 9: One Threat

Goal:

- add a simple hazard such as a patrolling drone or ground scavenger

Why it matters:

- introduces pressure and validates the intended tone

## Slice 10: Tighten And Refactor

Goal:

- clean up systems, naming, and docs only after the first playable loop exists

Why it matters:

- keeps early effort focused on discovering the game, not pre-optimizing structure
