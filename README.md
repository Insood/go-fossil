# Drone Fossil Hunter (go-fossil)

`go-fossil` is a game about recovering fossils from a junkyard. It is built with Go and [raylib](https://www.raylib.com/).

<img width="556" height="500" alt="covert-art" src="https://github.com/user-attachments/assets/3c500ccd-6379-42f5-b5a8-10f3bce69829" />

The player flies a drone over a desert junkyard, searches for artifacts, and then carefully cuts out artifacts from the ground while managing the drone battery. To gain points and to recharge the battery, the player must being the artifacts back to the spawn point. Once the player runs out of battery - the drone crashes and it is game over.

## Current state

This game was built for the 2026 ACM San Antonio [Velocicode Game Jam](https://acmsa.org/velocicode). Development will likely not continue after the game jam. However, the core loop is playable:

- Control a drone over terrain
- Aim and fire the cutting laser through the drone viewport
- Collect the fragments that were cut free
- Return fragments to the charging pad for score and battery re-charge
- Expand the map as the score grows.

## Controls

Keyboard and gamepad input are supported.

| Action | Keyboard and mouse | Gamepad |
| --- | --- | --- |
| Start the game | `Space` | A button |
| Move the drone | `W`, `A`, `S`, `D` | Left stick |
| Aim the cutting laser | Mouse over the drone viewport | Right stick |
| Fire the cutting laser | Left mouse button | Right trigger |
| Quit | Close the window | North face button |

## Build and run

The repository currently targets Go 1.26.4. On WSL/Ubuntu, build and run the
game with:

```bash
make run
```

To build without launching:

```bash
make build
```

The Linux build is written to `bin/linux/go-fossil`, with its runtime assets
copied to `bin/linux/assets` and `LICENSE.md` and `ASSETS.md` copied alongside
the executable.

To cross-compile a Windows build from the same environment:

```bash
make build-windows
```

This requires a MinGW-w64 cross-compiler and writes
`bin/win/go-fossil.exe` alongside `bin/win/assets`, `LICENSE.md`, and
`ASSETS.md`.

For a release build that does not open a console window:

```bash
make build-windows-release
```

Run the test suite with:

```bash
make test
```

## Documentation

- [Technical direction](docs/technical-direction.md) covers the architecture,
  gameplay systems, content formats, and implementation constraints.
- [LLM context](docs/llm-context.md) is the codebase map and working guide for
  coding agents.
