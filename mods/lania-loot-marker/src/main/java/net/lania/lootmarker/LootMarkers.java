package net.lania.lootmarker;

import it.unimi.dsi.fastutil.longs.Long2ByteMap;
import it.unimi.dsi.fastutil.longs.Long2ByteOpenHashMap;
import it.unimi.dsi.fastutil.longs.Long2ObjectMap;
import it.unimi.dsi.fastutil.longs.Long2ObjectOpenHashMap;
import net.minecraft.client.multiplayer.ClientLevel;
import net.minecraft.core.BlockPos;
import net.minecraft.world.level.ChunkPos;
import net.minecraft.world.level.block.ChestBlock;
import net.minecraft.world.level.block.state.BlockState;
import net.minecraft.world.level.block.state.properties.ChestType;

/**
 * Loot containers known for the current level: chunk -> block -> state. The server sends only the half of a double
 * chest that holds the loot, the other half is found from the chest block state. Belongs to one level: any access
 * with another level starts from empty, so a dimension change or reconnect never shows old markers.
 */
final class LootMarkers {

  private final Long2ObjectMap<Long2ByteMap> chunks = new Long2ObjectOpenHashMap<>();
  private ClientLevel level;

  void set(ClientLevel level, LootMarkersPayload payload) {
    use(level);
    long chunk = ChunkPos.pack(payload.chunkX(), payload.chunkZ());
    if (payload.positions().length == 0) {
      chunks.remove(chunk);
      return;
    }
    Long2ByteMap markers = new Long2ByteOpenHashMap(payload.positions(), payload.states());
    chunks.put(chunk, markers);
  }

  void dropChunk(ClientLevel level, ChunkPos pos) {
    if (level == this.level) {
      chunks.remove(pos.pack());
    }
  }

  void clear() {
    chunks.clear();
    level = null;
  }

  /** The state of the loot container at {@code pos}, either half of a double chest; -1 if it is not one. */
  byte stateAt(ClientLevel level, BlockPos pos) {
    use(level);
    byte state = stored(pos);
    if (state != -1) {
      return state;
    }
    BlockPos other = otherHalf(level, pos);
    return other == null ? -1 : stored(other);
  }

  void forEach(ClientLevel level, MarkerConsumer consumer) {
    use(level);
    for (Long2ByteMap markers : chunks.values()) {
      for (Long2ByteMap.Entry marker : markers.long2ByteEntrySet()) {
        consumer.accept(BlockPos.of(marker.getLongKey()), marker.getByteValue());
      }
    }
  }

  /** The connected half of a double chest, or null. */
  static BlockPos otherHalf(ClientLevel level, BlockPos pos) {
    BlockState block = level.getBlockState(pos);
    if (!(block.getBlock() instanceof ChestBlock) || block.getValue(ChestBlock.TYPE) == ChestType.SINGLE) {
      return null;
    }
    return pos.relative(ChestBlock.getConnectedDirection(block));
  }

  private byte stored(BlockPos pos) {
    Long2ByteMap markers = chunks.get(ChunkPos.pack(pos));
    return markers == null ? -1 : markers.getOrDefault(pos.asLong(), (byte) -1);
  }

  private void use(ClientLevel level) {
    if (level != this.level) {
      chunks.clear();
      this.level = level;
    }
  }

  interface MarkerConsumer {
    void accept(BlockPos pos, byte state);
  }
}
