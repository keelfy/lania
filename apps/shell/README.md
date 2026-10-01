# Shell

gRPC service that is the API's only gateway to the Minecraft server and its
plugins. The API never reads plugin schemas directly; it calls shell with
domain-level requests and shell translates them to plugin storage.

| Service             | Backed by                       |
|---------------------|---------------------------------|
| `PlayerService`     | Plan (playtime), Flectone (online status) |
| `PermissionService` | LuckPerms (groups, chat prefix). Groups are read from the database. A prefix change is an RCON command (`lp user <uuid> meta clear prefix`, then `meta addprefix`). A role change is written to the database, then the running server reloads it over RCON (`lp sync`) |
| `WhitelistService`  | Server whitelist, changed with RCON commands (`whitelist add/remove` by default, see `WHITELIST_ADD_COMMAND` and `WHITELIST_REMOVE_COMMAND` for whitelist plugins) |
| `AuthService`       | NavAuth on the proxy (in-game passwords of unlicensed players). Its tables live in their own database on the same MySQL server (`NAVAUTH_DATABASE_NAME`). The API sends a bcrypt hash; NavAuth reads credentials on every login, so a new password applies on the next login |
| `SkinService`       | SkinsRestorer on the proxy. Skins are read from its tables (`SKINSRESTORER_TABLE_PREFIX`, `sr_` by default) and changed with its console commands over RCON (`SKIN_SET_COMMAND`, `SKIN_CLEAR_COMMAND`). SkinsRestorer runs them in background, so shell polls `sr_players` until the change shows up, for 15 seconds at most |
| `PunishmentService` | LiteBans on the proxy. Bans and mutes are read from its tables (`LITEBANS_TABLE_PREFIX`, `litebans_` by default). Bans by IP alone are left out: they have no player |

Contracts live in `/proto/lania/shell/v1`. After editing them run
`mise run proto-generate`, which regenerates Go code for both shell and API.

## Layout

- `cmd` — entrypoint and wire setup
- `internal/storage` — plugin adapters, the only place that knows plugin schemas
- `internal/services` — domain logic (defaults for unknown players, prefix format)
- `internal/transport/rpc` — gRPC handlers, bearer-token auth, health check

## Running

```bash
cp .env.example .env
mise run shell-run
mise run shell-test
```

RCON is required for the whitelist and for prefixes: without `RCON_ADDRESS`, or
when the server is down, `WhitelistService` and `SetPlayerPrefix` calls fail.
For role changes it is optional: without it, or when the server is down, the
change stays in the database and applies on the next sync or restart; the RPC
still succeeds.

When `SHELL_TOKEN` is set, clients must send `authorization: Bearer <token>`.
The standard `grpc.health.v1.Health` service is exempt from auth.
