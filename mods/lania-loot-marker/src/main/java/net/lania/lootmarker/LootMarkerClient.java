package net.lania.lootmarker;

import net.fabricmc.api.ClientModInitializer;
import net.fabricmc.fabric.api.client.event.lifecycle.v1.ClientChunkEvents;
import net.fabricmc.fabric.api.client.event.lifecycle.v1.ClientTickEvents;
import net.fabricmc.fabric.api.client.networking.v1.ClientPlayConnectionEvents;
import net.fabricmc.fabric.api.client.networking.v1.ClientPlayNetworking;
import net.fabricmc.fabric.api.event.player.AttackBlockCallback;
import net.fabricmc.fabric.api.networking.v1.PayloadTypeRegistry;
import net.minecraft.client.Minecraft;
import net.minecraft.client.multiplayer.ClientLevel;
import net.minecraft.core.BlockPos;
import net.minecraft.core.particles.DustParticleOptions;
import net.minecraft.network.chat.Component;
import net.minecraft.util.RandomSource;
import net.minecraft.world.InteractionResult;

public final class LootMarkerClient implements ClientModInitializer {

  private static final DustParticleOptions CAN_LOOT = new DustParticleOptions(0xFFC400, 1.0F);
  private static final DustParticleOptions LOOTED = new DustParticleOptions(0x7F9CC0, 1.0F);
  private static final int PARTICLE_INTERVAL_TICKS = 10;
  private static final int PARTICLES_PER_MARKER = 2;
  private static final double PARTICLE_DISTANCE_SQR = 16 * 16;

  private final LootMarkers markers = new LootMarkers();
  private int ticks;

  @Override
  public void onInitializeClient() {
    // Registering the receiver makes the client announce the channel, which is what the server plugin waits for.
    PayloadTypeRegistry.clientboundPlay().register(LootMarkersPayload.TYPE, LootMarkersPayload.CODEC);
    ClientPlayNetworking.registerGlobalReceiver(LootMarkersPayload.TYPE, (payload, context) -> {
      ClientLevel level = context.client().level;
      if (level != null) {
        markers.set(level, payload);
      }
    });
    ClientChunkEvents.CHUNK_UNLOAD.register((level, chunk) -> markers.dropChunk(level, chunk.getPos()));
    ClientPlayConnectionEvents.DISCONNECT.register((handler, client) -> markers.clear());
    ClientTickEvents.END_CLIENT_TICK.register(this::spawnParticles);

    // Mirrors JustLootIt on the server: a loot container breaks only while sneaking.
    AttackBlockCallback.EVENT.register((player, level, hand, pos, direction) -> {
      if (!(level instanceof ClientLevel clientLevel) || player.isShiftKeyDown() || markers.stateAt(clientLevel, pos) == -1) {
        return InteractionResult.PASS;
      }
      player.sendOverlayMessage(Component.translatable("lania_loot_marker.sneak_to_break"));
      return InteractionResult.FAIL;
    });
  }

  private void spawnParticles(Minecraft client) {
    ClientLevel level = client.level;
    if (level == null || client.player == null || client.isPaused() || ++ticks % PARTICLE_INTERVAL_TICKS != 0) {
      return;
    }
    RandomSource random = level.getRandom();
    markers.forEach(level, (pos, state) -> {
      if (pos.distToCenterSqr(client.player.position()) > PARTICLE_DISTANCE_SQR) {
        return;
      }
      // A double chest gets particles around both halves.
      BlockPos other = LootMarkers.otherHalf(level, pos);
      double minX = pos.getX();
      double minZ = pos.getZ();
      double sizeX = 1;
      double sizeZ = 1;
      if (other != null) {
        minX = Math.min(minX, other.getX());
        minZ = Math.min(minZ, other.getZ());
        sizeX += Math.abs(other.getX() - pos.getX());
        sizeZ += Math.abs(other.getZ() - pos.getZ());
      }
      DustParticleOptions particle = state == LootMarkersPayload.LOOTED ? LOOTED : CAN_LOOT;
      for (int i = 0; i < PARTICLES_PER_MARKER; i++) {
        double x = minX - 0.1 + random.nextDouble() * (sizeX + 0.2);
        double y = pos.getY() + 0.1 + random.nextDouble() * 0.9;
        double z = minZ - 0.1 + random.nextDouble() * (sizeZ + 0.2);
        level.addParticle(particle, x, y, z, 0, 0, 0);
      }
    });
  }
}
