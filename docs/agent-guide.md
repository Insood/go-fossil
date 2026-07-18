# Agent Guide

This repository is developed incrementally with coding agents. The goal of this document is to keep changes coherent while the playable slice grows through small contributions.

## Working Style

- Make very small changes.
- Prefer vertical slices over broad scaffolding.
- Leave the project runnable after each meaningful step when possible.
- Document assumptions when they affect architecture, terrain behavior, camera behavior, system ownership, or gameplay rules.

## Current Playable Slice

The current slice includes:

- a raylib application shell and game loop
- an orthographic 3D main camera with dead-zone following
- heightmap-derived terrain chunks
- an ECS world with drone, light, and terrain chunk entities
- keyboard and gamepad drone movement
- a downward drone viewport with reticle aiming
- laser burns painted onto terrain overlay state
- embedded artifact overlays and per-pixel artifact ID masks
- periodic artifact cutout detection
- scored artifact fragments saved to disk and shown in UI
- generated chunk spawning after score increases
- shadow mapping plus debug tuning overlays

## Definition Of A Small Slice

A good slice changes one thing that can be observed or verified, such as:

- adjusting drone control
- changing camera follow behavior
- improving terrain generation
- changing artifact placement or scoring
- refining laser burn behavior
- improving fragment display
- tuning shadow rendering

Avoid slices that only create abstractions with no immediate behavior unless they unblock a concrete change.

## Priorities For Agents

When choosing between options, prefer:

1. visible progress
2. low coupling
3. reversible decisions
4. simple code
5. documented assumptions

## Expectations For Code Changes

- Keep naming concrete and gameplay-oriented.
- Do not introduce large frameworks beyond the stated stack unless requested.
- Avoid speculative systems for progression, inventory, enemies, hazards, AI, or content pipelines.
- Keep public APIs small.
- Use `cmd/game` for the executable and most ECS code.
- Keep one system per file.
- Only add `internal/*` packages when they own real game rules instead of convenience wrappers.

## Change Placement

- Put ECS components in `cmd/game/components.go`.
- Put one system per file in `cmd/game/*_system.go`.
- Put constants and tuning values in `cmd/game/config.go`.
- Put bootstrapping, system order, and game-owned asset lifecycle in `cmd/game/game.go`.
- Put asset loading details in `cmd/game/asset_manager.go`.
- Keep `main.go` as small as possible.
- Put terrain chunk parsing, validation, mesh data generation, baked terrain texture generation, and terrain sampling helpers in `internal/terrain`.
- Keep artifact runtime records, overlays, masks, scoring, and fragments in `cmd/game/artifact*.go` until those rules need a clearer boundary.

## Expectations For Documentation

When a code change introduces a meaningful architectural decision, update docs in the same change if practical.

Examples:

- terrain mutation strategy
- ECS component naming conventions
- camera behavior assumptions
- coordinate system conventions
- system ownership or ordering
- gameplay scoring rules

## Coordination Rules

Agents should explicitly note:

- what changed
- how the result can be verified
- any tracked files intentionally left untouched

## Do Not

- Do not move core gameplay logic into `main.go`.
- Do not create extra packages for speculative architecture.
- Do not add new dependencies unless they solve a real problem.
- Do not change core coordinate, camera, or terrain assumptions without updating docs.
- Do not replace the current texture/mask-based cutting loop with a larger terrain architecture unless requested.

## First Principle

If a task can be solved with a temporary, clear, playable implementation, prefer that over a larger architecture that delays feedback.
