package net.lania.lootmarker;

import java.io.ByteArrayOutputStream;
import java.io.DataOutputStream;
import java.io.IOException;
import java.io.UncheckedIOException;
import org.bukkit.Chunk;
import org.bukkit.block.BlockState;
import org.bukkit.block.Container;
import org.bukkit.entity.Player;
import org.bukkit.plugin.Plugin;

/**
 * Sends the loot containers of a chunk as seen by one player. The message replaces everything the client knows
 * about that chunk, so any change is a resend of the whole chunk.
 *
 * <p>Wire (big endian): {@code int chunkX, int chunkZ, varint n, n x (int x, int y, int z, byte state)},
 * state 0 = can loot, 1 = already looted.
 */
final class MarkerSender {

  static final String CHANNEL = "lania:loot_markers";

  private final Plugin plugin;
  private final JliBridge jli;

  MarkerSender(Plugin plugin, JliBridge jli) {
    this.plugin = plugin;
    this.jli = jli;
  }

  boolean listens(Player player) {
    return player.getListeningPluginChannels().contains(CHANNEL);
  }

  /** {@code evenIfEmpty = false} for a fresh chunk: the client has nothing to clear there. */
  void send(Player player, Chunk chunk, boolean evenIfEmpty) {
    if (!listens(player)) {
      return;
    }
    ByteArrayOutputStream entries = new ByteArrayOutputStream();
    DataOutputStream out = new DataOutputStream(entries);
    int count = 0;
    try {
      for (BlockState state : chunk.getTileEntities(false)) {
        if (!(state instanceof Container container)) {
          continue;
        }
        long id = jli.id(container.getPersistentDataContainer());
        if (id == -1) {
          continue;
        }
        out.writeInt(state.getX());
        out.writeInt(state.getY());
        out.writeInt(state.getZ());
        out.writeByte(jli.canLoot(chunk.getWorld(), id, player.getUniqueId()) ? 0 : 1);
        count++;
      }
      if (count == 0 && !evenIfEmpty) {
        return;
      }
      ByteArrayOutputStream message = new ByteArrayOutputStream(entries.size() + 13);
      DataOutputStream head = new DataOutputStream(message);
      head.writeInt(chunk.getX());
      head.writeInt(chunk.getZ());
      writeVarInt(head, count);
      entries.writeTo(message);
      player.sendPluginMessage(plugin, CHANNEL, message.toByteArray());
    } catch (IOException e) {
      throw new UncheckedIOException(e);
    }
  }

  private static void writeVarInt(DataOutputStream out, int value) throws IOException {
    while ((value & ~0x7F) != 0) {
      out.writeByte((value & 0x7F) | 0x80);
      value >>>= 7;
    }
    out.writeByte(value);
  }
}
