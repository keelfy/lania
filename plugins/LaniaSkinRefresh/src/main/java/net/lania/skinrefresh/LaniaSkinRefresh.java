package net.lania.skinrefresh;

import com.google.inject.Inject;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.event.player.ServerPostConnectEvent;
import com.velocitypowered.api.plugin.Plugin;
import com.velocitypowered.api.proxy.Player;
import com.velocitypowered.api.proxy.ProxyServer;
import java.util.concurrent.TimeUnit;
import org.slf4j.Logger;

@Plugin(id = "laniaskinrefresh", name = "LaniaSkinRefresh", version = "1.0.0",
    description = "Runs /skin update <player> from the console 5 seconds after a player connects to a server", authors = "keelfy")
public final class LaniaSkinRefresh {

  private final ProxyServer server;
  private final Logger logger;

  @Inject
  public LaniaSkinRefresh(ProxyServer server, Logger logger) {
    this.server = server;
    this.logger = logger;
  }

  // SkinsRestorer sometimes applies a stale skin after a server switch; a delayed update fixes it.
  // The command lives on the proxy, so it is run here rather than on the backend.
  @Subscribe
  public void onServerPostConnect(ServerPostConnectEvent event) {
    Player player = event.getPlayer();
    server.getScheduler().buildTask(this, () -> {
      if (!player.isActive()) {
        return;
      }
      String command = "skin update " + player.getUsername();
      server.getCommandManager().executeAsync(server.getConsoleCommandSource(), command).thenAccept(known ->
          logger.info("/{}{}", command, known ? "" : " -> command not found"));
    }).delay(5, TimeUnit.SECONDS).schedule();
  }
}
