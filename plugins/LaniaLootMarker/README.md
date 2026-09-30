# LaniaLootMarker

A Paper plugin that tells the `lania-loot-marker` client mod (`mods/lania-loot-marker`) where the JustLootIt loot
containers are and whether the player has already looted each one. The mod draws particles and asks to sneak
before breaking. Players without the mod get nothing.

## How it works

- For every chunk sent to a player with the mod, the plugin sends the containers that have a JustLootIt id, each
  with "can loot" or "looted" for that player, over the `lania:loot_markers` plugin channel.
- One message holds one whole chunk, and the client replaces what it knew about that chunk.
- After a player opens a container, the plugin resends the chunk to that player on the next tick.
- After a container is broken or blown up, the plugin resends the chunk to everyone who sees it.
- Both resends wait a tick because JustLootIt cancels these events and changes the container itself.
- A refresh timer that runs out does not update the marker right away. The marker changes the next time the chunk
  is sent to the player.

Wire (big endian): `int chunkX, int chunkZ, varint n, n × (int x, int y, int z, byte state)`, where state 0 means
the player can loot and 1 means already looted.

## JustLootIt version

JustLootIt has no stable API, so `JliBridge` reads its internals directly. The version in `build.gradle` must
match the jar on the server. After a JustLootIt update, rebuild against the new version and check `JliBridge`.

## Install

1. Build: `./gradlew build` (Java 25), the jar is `build/libs/LaniaLootMarker-<version>.jar`.
2. Put the jar into `plugins` next to JustLootIt and restart. There is no config.
