# LaniaSkinRefresh

A Velocity plugin that runs `skin update <player>` from the console 5 seconds after a player connects to any server in the network.
SkinsRestorer sometimes shows a wrong skin after switching servers, and this update fixes it.

The `/skin` command is registered by SkinsRestorer on the proxy, so this plugin has to run on Velocity too.

## Install

1. Build: `./gradlew build` (Java 25 toolchain, targets Java 21), the jar is `build/libs/LaniaSkinRefresh-<version>.jar`.
2. Put the jar into the Velocity `plugins` folder and restart the proxy. There is no config.
