# Lania Admin

Standalone Go HTTP service for existing Kratos accounts and Minecraft profiles.
It serves its own HTML interface and talks directly to MariaDB, the private
Kratos admin API, and `shell`. It makes no requests to the public Lania API.
The binary shares the API's Go module and domain code to preserve product
metadata and Minecraft formatting. Its entry point is `apps/api/cmd/admin`;
it has its own image, process, connection pool (maximum five), and memory limit.
MariaDB remains shared, so database load is not isolated.

## Capabilities

- Browse accounts, including accounts without profiles; exact email search and pagination.
- Find profiles by nickname or UUID; view profiles belonging to an account.
- Attach an existing profile to an existing account or detach it. To transfer a
  profile, detach it first. A stale owner form returns a conflict.
- Grant season access, name colors, and name prefixes for a chosen season.
  Repeated grants do not duplicate existing rights. Granting cosmetics does not
  select them automatically.
- Revoke individual cosmetic rights, including permanent rights, or all access
  rights for the selected season. The default color is protected. Selected
  cosmetics reset only when no permanent or current-season right remains.
- Synchronize whitelist membership and selected cosmetics through `shell`.
  Whitelist removal prevents subsequent joins; it does not kick an online player.

This service does not create accounts, profiles, products, orders, or refunds.
Revoking a product preserves the original order history. Profile limits used by
public checkout do not apply to administrative attachment.

## Authentication

Use existing Kratos browser sessions. `ADMIN_IDENTITY_IDS` is a comma-separated
allowlist of Kratos identity UUIDs, not profile UUIDs. An empty list denies all
access. The service checks session activity and the allowlist on every request.
It does not use user-editable traits, Minecraft roles, or frontend checks to
assign administrator privileges.

The production session cookie already covers `lania.network`. Kratos allows
`https://admin.lania.network` as a login return URL. Login continues through the
existing website's login UI. The Kratos admin API remains private.

All changes require POST with an `Origin` exactly matching `ADMIN_ORIGIN`.
The interface uses browser forms; requests without Origin are rejected.
Responses cannot be cached or embedded in frames. No credential fields from
Kratos are passed to the UI. Mutation logs contain actor, profile, action,
product/grant, season, and owner IDs.

## Configuration and launch

Set these environment variables before `mise run admin-run`:

| Variable | Meaning |
| --- | --- |
| `PORT` | Listen port, default `8081` |
| `ADMIN_ORIGIN` | Browser origin, e.g. `https://admin.lania.network` |
| `ADMIN_IDENTITY_IDS` | Comma-separated administrator identity UUIDs |
| `ORY_URL` | Internal Kratos public API, e.g. `http://kratos:4433` |
| `ORY_ADMIN_URL` | Internal Kratos admin API, e.g. `http://kratos:4434` |
| `ORY_BROWSER_URL` | Browser-accessible Kratos public API |
| `DATABASE_HOST`, `DATABASE_PORT` | MariaDB host and port (default `3306`) |
| `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME` | Existing application database |
| `ACTIVE_SEASON_ID`, `DEFAULT_NAME_COLOR_ID` | Same values as the API |
| `SHELL_ADDRESS`, `SHELL_TOKEN` | Same shell endpoint and token as the API |

For local development, use a hostname covered by the existing Kratos cookie and
add that admin origin to Kratos `selfservice.allowed_return_urls`. A different
localhost origin does not receive cookies scoped to `dev.lania.network`.
If using a reverse proxy, forward admin requests without changing their Origin.

```bash
mise run admin-build
mise run admin-test
docker build -f apps/admin/Dockerfile -t lania-admin .
```

## Production

Set `admin_identity_ids` in Terraform to the existing accounts that should have
access. Apply the web stack to create the `admin` DNS record, update the Kratos
return URL, and install the new compose service. There are no new database
migrations. The service waits for the existing migration job.

CI publishes `ghcr.io/keelfy/lania:latest-admin` and `admin-<commit SHA>` tags.
`admin_image_tag` / `ADMIN_IMAGE_TAG` can pin an admin image independently of the
other services for rollback. Open `https://admin.lania.network` after deployment.
The admin image uses the registry tags emitted by the current workflow.

## Minecraft failures

Database changes commit before Minecraft synchronization. A failed RPC produces
an explicit warning on the profile page. Use **Repeat synchronization** after
shell becomes available; retries read current database state under a profile
lock. Synchronization is synchronous and manually retried, with no background
queue. If the service stops after committing but before calling shell, use the
same button. A process restart alone does not retry pending changes.

## Verification

Unit tests cover authentication, authorization, cross-origin request rejection,
pagination, credential projection, and escaped HTML templates. Integration tests
use the repository's initial MariaDB schema and exercise product rights,
selected cosmetics, ownership conflicts, and shell failure/retry behavior.

```bash
cd apps/api
# Use a disposable MariaDB server. Tests create and drop unique admin_test_* databases.
CGO_ENABLED=1 ADMIN_TEST_DSN='root@tcp(127.0.0.1:3306)/' go test -race ./internal/admin
```

Without `ADMIN_TEST_DSN`, integration tests are skipped. `/healthz` is an
unauthenticated liveness endpoint; it does not claim dependency readiness.
