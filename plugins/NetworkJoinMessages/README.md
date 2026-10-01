## About

This is a continuation of Tirco's project [BungeeJoinMessages](https://github.com/Tirco/BungeeJoinMessages) as it appears it is no longer being maintained.

## Build

Use Java 21 and run `mvn --batch-mode verify`. The installable JAR is written
to `target/NetworkJoinMessages-2.3.2.jar`.

The Maven configuration is restored from the
[upstream project](https://github.com/lania-smp/NetworkJoinMessages/blob/master/pom.xml),
with Maven Central preferred, snapshots limited to Paper and Elytrium,
MiniPlaceholders 3 matching the source API, and dependency-reduced POM generation
disabled.

## License

Zlib was chosen as the basis for this project (BukkitPlugin) as it is highly permissive and easy for people to understand. The license has only been modified for this project to reflect authorship and creation year.

Copyright (c) 2021 Tirco and EarthCow

This software is provided 'as-is', without any express or implied
warranty. In no event will the authors be held liable for any damages
arising from the use of this software.

Permission is granted to anyone to use this software for any purpose,
including commercial applications, and to alter it and redistribute it
freely, subject to the following restrictions:

1. The origin of this software must not be misrepresented; you must not
   claim that you wrote the original software. If you use this software
   in a product, an acknowledgment in the product documentation would be
   appreciated but is not required.

2. Altered source versions must be plainly marked as such, and must not be
   misrepresented as being the original software.

3. This notice may not be removed or altered from any source
   distribution.
