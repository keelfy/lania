package net.lania.verify;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.file.Path;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.TimeUnit;

import org.slf4j.Logger;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.google.inject.Inject;
import com.velocitypowered.api.event.EventTask;
import com.velocitypowered.api.event.ResultedEvent.ComponentResult;
import com.velocitypowered.api.event.Subscribe;
import com.velocitypowered.api.event.connection.LoginEvent;
import com.velocitypowered.api.event.proxy.ProxyInitializeEvent;
import com.velocitypowered.api.plugin.Plugin;
import com.velocitypowered.api.plugin.annotation.DataDirectory;
import com.velocitypowered.api.proxy.Player;

import net.kyori.adventure.text.minimessage.MiniMessage;

/**
 * Asks the Lania API on every licensed login whether the owner of the profile waits for a verification code, and
 * shows the code on the kick screen. The owner types it on the site: only the licensed player can see it.
 * Any failure lets the player in as usual.
 */
@Plugin(id = "laniaverify", name = "LaniaVerify", version = "1.0.0",
    description = "License verification codes for lania.network", authors = {"keelfy"})
public final class LaniaVerify {

  private static final Gson GSON = new Gson();

  private final Logger logger;
  private final Path dataDirectory;
  private final HttpClient http = HttpClient.newHttpClient();
  private volatile Settings settings;

  @Inject
  public LaniaVerify(Logger logger, @DataDirectory Path dataDirectory) {
    this.logger = logger;
    this.dataDirectory = dataDirectory;
  }

  @Subscribe
  public void onProxyInitialization(ProxyInitializeEvent event) {
    try {
      settings = Settings.load(dataDirectory);
    } catch (Exception e) {
      logger.error("Failed to load config.properties, verification is off", e);
    }
  }

  // The lowest priority runs last, so the code screen replaces a denial of a whitelist that ran before:
  // a player waiting for the code does not need access to the server.
  @Subscribe(priority = Short.MIN_VALUE)
  public EventTask onLogin(LoginEvent event) {
    Settings current = settings;
    Player player = event.getPlayer();
    // Only a login checked by Mojang proves the license.
    if (current == null || !current.enabled() || !player.isOnlineMode()) {
      return null;
    }
    return EventTask.resumeWhenComplete(requestCode(current, player)
        .thenAccept(code -> {
          if (code != null && !code.isEmpty()) {
            String message = current.message().replace("{code}", code);
            event.setResult(ComponentResult.denied(MiniMessage.miniMessage().deserialize(message)));
          }
        })
        .exceptionally(e -> {
          logger.warn("Verification check of {} failed, letting the login go on: {}", player.getUsername(), e.toString());
          return null;
        }));
  }

  private CompletableFuture<String> requestCode(Settings current, Player player) {
    JsonObject body = new JsonObject();
    body.addProperty("uuid", player.getUniqueId().toString());
    body.addProperty("username", player.getUsername());
    body.addProperty("onlineMode", player.isOnlineMode());

    HttpRequest request = HttpRequest.newBuilder(URI.create(current.apiUrl() + "/v1/internal/verification/login"))
        .timeout(current.timeout())
        .header("Content-Type", "application/json")
        .header("x-api-key", current.apiKey())
        .POST(HttpRequest.BodyPublishers.ofString(GSON.toJson(body)))
        .build();
    return http.sendAsync(request, HttpResponse.BodyHandlers.ofString())
        .orTimeout(current.timeout().toMillis(), TimeUnit.MILLISECONDS)
        .thenApply(response -> {
          if (response.statusCode() != 200) {
            throw new IllegalStateException("API answered " + response.statusCode());
          }
          JsonObject answer = GSON.fromJson(response.body(), JsonObject.class);
          return answer != null && answer.has("code") ? answer.get("code").getAsString() : null;
        });
  }
}
