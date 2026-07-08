# Agent Guide

This repository will be developed incrementally with coding agents. The goal of this document is to keep work coherent even when many small contributions happen over time.

## Working Style

- Make very small changes.
- Prefer vertical slices over broad scaffolding.
- Leave the project runnable after each meaningful step when possible.
- Document assumptions when they affect later architecture.

## Definition Of A Small Slice

A good slice usually changes one thing that can be observed or verified, such as:

- opening a window
- drawing terrain
- spawning the player drone
- moving the drone
- drawing a salvageable artifact
- cutting a visible chunk of terrain

Avoid slices that only create abstractions with no immediate behavior unless they unblock the next concrete step.

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
- Avoid speculative systems for progression, inventory, AI, or content pipelines.
- Keep public APIs small until real usage patterns emerge.
- Default to the `go-towerdefense` repo shape.
- Use `cmd/game` for the executable and most ECS code.
- Keep one system per file.
- Only add `internal/*` packages when they own real game rules instead of convenience wrappers.

## Change Placement

- Put ECS components in `cmd/game/components.go`.
- Put one system per file in `cmd/game/*_system.go`.
- Put constants and tuning values in `cmd/game/config.go`.
- Put bootstrapping, system order, and asset lifecycle in `cmd/game/game.go`.
- Keep `main.go` as small as possible.
- Put isolated domain logic in `internal/*` only when the package boundary is clearly useful.

## Expectations For Documentation

When a code change introduces a meaningful architectural decision, update docs in the same change if practical.

Examples:

- chosen terrain mutation strategy
- ECS component naming conventions
- camera behavior assumptions
- coordinate system conventions

## Coordination Rules

Agents should explicitly note:

- what assumption they made
- what remains incomplete
- how the result can be verified

This matters because future work will likely continue from partial prototypes rather than finished systems.

## Do Not

- Do not move core gameplay logic into `main.go`.
- Do not create extra packages for speculative architecture.
- Do not add new dependencies unless they solve a real problem.
- Do not change core coordinate, camera, or terrain assumptions without updating docs.

## First Principle

If a task can be solved with a temporary, clear, playable implementation, prefer that over a larger “correct” architecture that delays feedback.
