package net.lania.lootmarker;

import org.bukkit.plugin.java.JavaPlugin;

public final class LaniaLootMarker extends JavaPlugin {

  @Override
  public void onEnable() {
    getServer().getMessenger().registerOutgoingPluginChannel(this, MarkerSender.CHANNEL);
    MarkerSender sender = new MarkerSender(this, new JliBridge());
    getServer().getPluginManager().registerEvents(new MarkerListener(this, sender), this);
  }
}
