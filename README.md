# Battle Worlds

## Enterprise edition

We build and deploy a paid Enterprise edition to order. Examples of what we can develop and extend for your organization are listed below.

- Internet multiplayer, custom open worlds and persistent game servers scoped per project.
- Central control for server configuration, map versions and events across hosted worlds.
- Player, moderator and administrator roles, with integrations for private communities and event systems.

For an Enterprise edition or custom development, contact us at [verobots.eu](https://verobots.eu/) or [olarian.com](https://olarian.com/).


Explore seven themed 3D worlds, unlock new abilities and ride a flying train between islands. The game runs locally and supports single-player exploration and combat.

## Play on Windows

Download the Windows x64 ZIP from Releases, extract it to a folder and open **battle-worlds.exe**. An OpenGL 3.3-capable graphics driver is required. The interface is in Romanian.

Use **WASD** to move, the mouse to look, **Space** to jump and **Esc** to open the pause menu. Follow the on-screen instructions to select an island and unlock its next world.

## Build

Run `go test ./...` and `go build -trimpath -o battle-worlds.exe .` with Go 1.26 or newer on Windows. Audio assets are embedded. See [audio credits](assets/sfx/CREDITS.txt) for their sources and terms.

Source licensed under [MIT](LICENSE). The game saves progress in the current user's application data directory.
