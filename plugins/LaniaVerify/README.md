# LaniaVerify

A Velocity plugin for the license checkmark on lania.network. The owner of a profile starts the verification on
the site; when the licensed player joins, the proxy kicks them with a code, and the owner types it on the site.
Only the licensed player can see the code, which proves the account is theirs.

## How it works

- On every login checked by Mojang (`Player#isOnlineMode`) the plugin posts `uuid`, `username`, `onlineMode` to
  `POST {apiUrl}/v1/internal/verification/login` with the `x-api-key` header.
- The API answers `{"code": "..."}` when the profile has an open verification, `{}` otherwise.
- A code replaces the login result with a kick that shows `message`. The handler runs last, so the code is shown
  even when a whitelist denied the player before.
- No code, an error or no answer within `timeoutMs` leaves the login untouched.
- Offline (non-licensed) logins are not sent to the API.

## Install

1. Build: `./gradlew build` (Java 21), the jar is `build/libs/LaniaVerify-<version>.jar`.
2. Put the jar into the `plugins` directory of the proxy and start it once.
3. Edit `plugins/laniaverify/config.properties`: `apiKey` must equal `API_KEY` of the API. Restart the proxy.

## Config

| Key | Default | Meaning |
|---|---|---|
| `enabled` | `true` | `false` turns the plugin off |
| `apiUrl` | `https://api.lania.network` | Lania API base URL |
| `apiKey` | `CHANGEME` | `API_KEY` of the API |
| `timeoutMs` | `1500` | how long a login waits for the API |
| `message` | see the file | kick screen in MiniMessage, `{code}` is the code |
