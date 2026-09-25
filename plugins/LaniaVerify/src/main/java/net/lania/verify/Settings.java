package net.lania.verify;

import java.io.IOException;
import java.io.InputStream;
import java.io.Reader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.time.Duration;
import java.util.Properties;

/** Settings of config.properties in the data directory, written with the defaults on the first start. */
record Settings(boolean enabled, String apiUrl, String apiKey, Duration timeout, String message) {

  static Settings load(Path dataDirectory) throws IOException {
    Path file = dataDirectory.resolve("config.properties");
    if (Files.notExists(file)) {
      Files.createDirectories(dataDirectory);
      try (InputStream defaults = Settings.class.getResourceAsStream("/config.properties")) {
        Files.copy(defaults, file);
      }
    }

    Properties properties = new Properties();
    try (Reader reader = Files.newBufferedReader(file, StandardCharsets.UTF_8)) {
      properties.load(reader);
    }
    return new Settings(
        Boolean.parseBoolean(properties.getProperty("enabled", "true")),
        properties.getProperty("apiUrl", "").replaceAll("/+$", ""),
        properties.getProperty("apiKey", ""),
        Duration.ofMillis(Long.parseLong(properties.getProperty("timeoutMs", "1500").trim())),
        properties.getProperty("message", "{code}"));
  }
}
