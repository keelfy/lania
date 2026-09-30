package net.lania.lootmarker;

import io.papermc.paper.event.packet.PlayerChunkLoadEvent;
import java.util.HashSet;
import java.util.List;
import java.util.Set;
import org.bukkit.Chunk;
import org.bukkit.block.Block;
import org.bukkit.block.BlockFace;
import org.bukkit.block.Container;
import org.bukkit.entity.Player;
import org.bukkit.event.EventHandler;
import org.bukkit.event.EventPriority;
import org.bukkit.event.Listener;
import org.bukkit.event.block.Action;
import org.bukkit.event.block.BlockBreakEvent;
import org.bukkit.event.block.BlockExplodeEvent;
import org.bukkit.event.entity.EntityExplodeEvent;
import org.bukkit.event.player.PlayerInteractEvent;
import org.bukkit.event.player.PlayerRegisterChannelEvent;
import org.bukkit.plugin.Plugin;

/**
 * JustLootIt cancels open and break events and changes containers by itself, so container changes are not read
 * from the events: the affected chunks are resent on the next tick, when JustLootIt is done.
 */
final class MarkerListener implements Listener {

  private static final BlockFace[] SIDES = {BlockFace.NORTH, BlockFace.SOUTH, BlockFace.EAST, BlockFace.WEST};

  private final Plugin plugin;
  private final MarkerSender sender;

  MarkerListener(Plugin plugin, MarkerSender sender) {
    this.plugin = plugin;
    this.sender = sender;
  }

  @EventHandler
  public void onChunkLoad(PlayerChunkLoadEvent event) {
    sender.send(event.getPlayer(), event.getChunk(), false);
  }

  /** The client registers the channel after the first chunks of the join are already sent. */
  @EventHandler
  public void onRegisterChannel(PlayerRegisterChannelEvent event) {
    if (!event.getChannel().equals(MarkerSender.CHANNEL)) {
      return;
    }
    Player player = event.getPlayer();
    for (Chunk chunk : player.getSentChunks()) {
      sender.send(player, chunk, false);
    }
  }

  /** Opening changes only the state of the player who opened. */
  @EventHandler(priority = EventPriority.MONITOR)
  public void onInteract(PlayerInteractEvent event) {
    Block block = event.getClickedBlock();
    if (event.getAction() != Action.RIGHT_CLICK_BLOCK || block == null || !isContainer(block)) {
      return;
    }
    Player player = event.getPlayer();
    if (!sender.listens(player)) {
      return;
    }
    Set<Chunk> chunks = chunksAround(block);
    plugin.getServer().getScheduler().runTask(plugin, () -> {
      if (player.isOnline()) {
        chunks.forEach(chunk -> sender.send(player, chunk, true));
      }
    });
  }

  @EventHandler(priority = EventPriority.MONITOR)
  public void onBreak(BlockBreakEvent event) {
    resendIfContainers(List.of(event.getBlock()));
  }

  @EventHandler(priority = EventPriority.MONITOR)
  public void onBlockExplode(BlockExplodeEvent event) {
    resendIfContainers(event.blockList());
  }

  @EventHandler(priority = EventPriority.MONITOR)
  public void onEntityExplode(EntityExplodeEvent event) {
    resendIfContainers(event.blockList());
  }

  private void resendIfContainers(List<Block> blocks) {
    Set<Chunk> chunks = new HashSet<>();
    for (Block block : blocks) {
      if (isContainer(block)) {
        chunks.addAll(chunksAround(block));
      }
    }
    if (chunks.isEmpty()) {
      return;
    }
    plugin.getServer().getScheduler().runTask(plugin, () -> {
      for (Chunk chunk : chunks) {
        if (chunk.isLoaded()) {
          chunk.getWorld().getPlayersSeeingChunk(chunk).forEach(player -> sender.send(player, chunk, true));
        }
      }
    });
  }

  private static boolean isContainer(Block block) {
    return block.getState(false) instanceof Container;
  }

  /** A double chest can cross a chunk border, and JustLootIt may move the id to the other half. */
  private static Set<Chunk> chunksAround(Block block) {
    Set<Chunk> chunks = new HashSet<>();
    chunks.add(block.getChunk());
    for (BlockFace side : SIDES) {
      int chunkX = (block.getX() + side.getModX()) >> 4;
      int chunkZ = (block.getZ() + side.getModZ()) >> 4;
      if (block.getWorld().isChunkLoaded(chunkX, chunkZ)) {
        chunks.add(block.getWorld().getChunkAt(chunkX, chunkZ));
      }
    }
    return chunks;
  }
}
