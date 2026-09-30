package net.lania.lootmarker;

import net.minecraft.core.BlockPos;
import net.minecraft.network.FriendlyByteBuf;
import net.minecraft.network.codec.StreamCodec;
import net.minecraft.network.protocol.common.custom.CustomPacketPayload;
import net.minecraft.resources.Identifier;

/**
 * All JustLootIt loot containers of one chunk, sent by the LaniaLootMarker server plugin. Replaces what the client
 * knew about the chunk. Wire: {@code int chunkX, int chunkZ, varint n, n x (int x, int y, int z, byte state)}.
 */
record LootMarkersPayload(int chunkX, int chunkZ, long[] positions, byte[] states) implements CustomPacketPayload {

  static final byte CAN_LOOT = 0;
  static final byte LOOTED = 1;

  static final Type<LootMarkersPayload> TYPE = new Type<>(Identifier.fromNamespaceAndPath("lania", "loot_markers"));
  static final StreamCodec<FriendlyByteBuf, LootMarkersPayload> CODEC =
      CustomPacketPayload.codec(LootMarkersPayload::write, LootMarkersPayload::read);

  private static LootMarkersPayload read(FriendlyByteBuf buf) {
    int chunkX = buf.readInt();
    int chunkZ = buf.readInt();
    int count = buf.readVarInt();
    long[] positions = new long[count];
    byte[] states = new byte[count];
    for (int i = 0; i < count; i++) {
      positions[i] = BlockPos.asLong(buf.readInt(), buf.readInt(), buf.readInt());
      states[i] = buf.readByte();
    }
    return new LootMarkersPayload(chunkX, chunkZ, positions, states);
  }

  private void write(FriendlyByteBuf buf) {
    buf.writeInt(chunkX);
    buf.writeInt(chunkZ);
    buf.writeVarInt(positions.length);
    for (int i = 0; i < positions.length; i++) {
      buf.writeInt(BlockPos.getX(positions[i]));
      buf.writeInt(BlockPos.getY(positions[i]));
      buf.writeInt(BlockPos.getZ(positions[i]));
      buf.writeByte(states[i]);
    }
  }

  @Override
  public Type<LootMarkersPayload> type() {
    return TYPE;
  }
}
