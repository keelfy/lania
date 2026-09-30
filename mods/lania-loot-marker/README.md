# lania-loot-marker

A client-only Fabric mod (Minecraft 26.3) for lania.network. It works together with the `LaniaLootMarker` server
plugin (`plugins/LaniaLootMarker`).

- Loot containers from JustLootIt get a few particles around them. Gold means you can loot the container, gray-blue
  means you already looted it.
- A loot container, including both halves of a double chest, breaks only while you sneak. Without sneaking, the
  attack is cancelled on the client and the action bar shows a hint. The server enforces the same rule; the mod
  just stops the break animation from starting.
- On servers without the plugin the mod does nothing.

## Build

`./gradlew build` (Java 25), the jar is `build/libs/lania-loot-marker-<version>.jar`. It needs Fabric API.
