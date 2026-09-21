# Shell

gRPC service that is the API's only gateway to the Minecraft server and its
plugins. The API never reads plugin schemas directly; it calls shell with
domain-level requests and shell translates them to plugin storage.

| Service             | Backed by                       |
|---------------------|---------------------------------|
| `PlayerService`     | Plan (playtime), Flectone (online status) |
| `PermissionService` | LuckPerms (groups, chat prefix). Groups are read from the database. A prefix change is an RCON command (`lp user <uuid> meta clear prefix`, then `meta addprefix`). A role change is written to the database, then the running server reloads it over RCON (`lp sync`) |
| `WhitelistService`  | Server whitelist, changed with RCON commands (`whitelist add/remove` by default, see `WHITELIST_ADD_COMMAND` and `WHITELIST_REMOVE_COMMAND` for whitelist plugins) |

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
