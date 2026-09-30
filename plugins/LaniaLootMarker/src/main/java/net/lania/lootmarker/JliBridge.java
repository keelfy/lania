package net.lania.lootmarker;

import java.util.UUID;
import me.lauriichan.spigot.justlootit.JustLootItAccess;
import me.lauriichan.spigot.justlootit.JustLootItPlugin;
import me.lauriichan.spigot.justlootit.capability.StorageCapability;
import me.lauriichan.spigot.justlootit.data.Container;
import me.lauriichan.spigot.justlootit.storage.Stored;
import org.bukkit.World;
import org.bukkit.persistence.PersistentDataContainer;

/**
 * The only place that touches JustLootIt. It has no stable API, so this follows the internals of the pinned
 * version (see build.gradle); a JustLootIt update may need changes here.
 */
final class JliBridge {

  private final JustLootItPlugin jli = JustLootItPlugin.get();

  /** The container id, or -1 when the block is not a JustLootIt container (the other half of a double chest too). */
  long id(PersistentDataContainer data) {
    return JustLootItAccess.hasIdentity(data) ? JustLootItAccess.getIdentity(data) : -1;
  }

  /** True when the container would give the player loot now. Unknown state (storage busy or missing) counts as true. */
  boolean canLoot(World world, long id, UUID player) {
    StorageCapability storage = jli.versionHandler().getLevel(world).getCapability(StorageCapability.class).orElse(null);
    if (storage == null || storage.hasBulkOperationRunning()) {
      return true;
    }
    Stored<Container> stored = storage.storage().read(id);
    return stored == null || stored.value().canAccess(world, player);
  }
}
